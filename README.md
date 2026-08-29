# micro-kit

Go 语言微服务基础组件库，聚焦于微服务通用基础设施。项目提供统一的配置加载、日志、链路追踪、监控、HTTP / gRPC 网关、数据库、缓存、消息队列、认证、错误处理等能力，适合作为业务系统的基础支撑层。

## 项目定位

本仓库采用“轻封装 + 可插拔配置 + 统一约定”的方式组织基础能力，核心原则包括：

- 统一配置入口，支持本地 YAML 与 Nacos 配置中心
- 统一日志和链路追踪接入
- 对常见中间件做薄封装，不强绑定业务代码
- 服务启动支持按需启用 Nacos 注册 / 服务发现
- 适合在 Kratos 微服务架构下使用

## 目录说明

- apps/kit: 应用初始化入口
- cloud/config: 配置初始化和统一读取入口
- cloud/register/nacos: Nacos 配置中心与服务注册发现封装
- cloud/server: 服务启动与自动注册能力
- cloud/tracer: OpenTelemetry 链路追踪
- cloud/metrics: 指标与监控
- clients: 各类基础客户端封装
  - ihttp: HTTP 客户端
  - iredis: Redis 客户端
  - igorm: MySQL / GORM
  - ikafka: Kafka 消费 / 生产封装
  - irocketmq: RocketMQ 消费 / 生产封装
  - ialiyun/ioss: Aliyun OSS 封装
  - igrpc/kratos: gRPC 客户端连接与服务发现
- middleware/kratos: Kratos HTTP / gRPC 中间件
- utils: 日志、JWT、上下文、环境变量、异步工具
- errors: 统一错误处理
- protos: protobuf 定义，承载基础配置结构

## 配置设计

项目采用 protobuf 结构定义配置项，并通过配置加载器读取：

- 本地配置文件：`configs/debug/config.yaml`
- 远端配置中心：Nacos
- 配置结构定义：`protos/protos.proto`

配置读取入口：

```go
cfg := config.Get("mysql.default")
// 或直接扫描到结构体
var db protos.Mysql
if err := config.Get("mysql.default").Scan(&db); err != nil {
    panic(err)
}
```

### Nacos 配置加载

项目支持两层配置加载：

1. 本地 bootstrap / 本地 YAML
2. Nacos 配置中心（可选增强）

启动时可以显式选择配置来源：

- `kit.WithLocalConfig()`: 仅加载本地 YAML
- `kit.WithNacosConfig(cfg)`: 读取 Nacos 配置中心
- 默认行为：优先读取本地配置，并在配置存在时补充/覆盖 Nacos 配置

正常生产环境中，推荐做法是：

- 本地保留最小 bootstrap 配置，用于连接 Nacos
- 由 Nacos 拉取实际运行配置
- 若 Nacos 不可用，则回退到本地配置，避免服务直接无法启动

示例：

```yaml
nacos:
  config_center:
    servers:
      - ip_addr: "127.0.0.1"
        port: 8848
    client:
      namespace_id: "public"
      timeout_ms: 5000
      log_level: "info"
    data_id: "micro-kit"
    group: "DEFAULT_GROUP"
```

```go
if err := kit.Init("my-service",
    kit.WithNacosConfig(&protos.NacosConfigCenter{
        DataId: "micro-kit",
        Group:  "DEFAULT_GROUP",
        Servers: []*protos.NacosServerConfig{{IpAddr: "127.0.0.1", Port: 8848}},
    }),
    kit.WithTracer(),
); err != nil {
    panic(err)
}
```

## 启动方式

### 1. 初始化应用

启动入口统一使用 `kit.Init`，并支持按需注册启动 hook：

```go
package main

import (
    "github.com/reaburoa/micro-kit/apps/kit"
)

func main() {
    if err := kit.Init("my-service",
        kit.WithTracer(),
        kit.WithLocalConfig(),
    ); err != nil {
        panic(err)
    }
}
```

启动顺序为：

1. 先应用所有 option，注册 hook / 配置来源
2. 调用 `config.InitConfig()` 初始化配置
3. 执行统一的 `runHooks()`，初始化日志、指标、链路追踪等依赖配置的组件
4. 启动退出信号监听和 shutdown 清理

这样可以确保像 tracing 这类依赖配置的模块一定在配置加载完成之后再初始化，避免启动时序错误。

### 2. HTTP 服务

```go
srv := kratos.NewHttp()
_ = srv
```

### 3. gRPC 服务

```go
gSrv := kratos.NewGrpc()
_ = gSrv
```

### 4. 服务注册与发现

项目支持在 HTTP / gRPC 启动时按配置自动注册到 Nacos。

默认行为是关闭的，需要显式打开：

```yaml
server:
  http:
    port: 8080
    enable_nacos: true
    enable_watch: false
```

这样在服务启动时，可选择自动注册到 Nacos，配合 gRPC / HTTP 客户端服务发现使用。

## 组件能力

### 日志

统一的日志初始化与 zap 封装，支持输出到 stdout 和文件，配置项在 `protos.Logger` 中。

### 链路追踪

基于 OpenTelemetry，支持：

- HTTP tracing
- gRPC tracing
- Kafka tracing
- RocketMQ tracing
- OSS tracing

### 监控

提供统一的 metrics 和 server 监控能力，适合和 Prometheus / OTEL 体系接入。

### Redis

```go
client, shutdown, err := iredis.RedisClient("default")
if err != nil {
    panic(err)
}
defer shutdown()
```

### MySQL / GORM

```go
db, shutdown, err := igorm.GormClient("default")
if err != nil {
    panic(err)
}
defer shutdown()
```

### Kafka

- 支持 producer / consumer 统一封装
- 可通过 `kafka.<topic>` 配置读取
- 支持 tracing 和消息属性透传

### RocketMQ

- 支持 producer / consumer 统一封装
- 使用 `protos.RocketMQ` 定义配置
- 支持批量拉取、消费组、tag 等参数配置

### OSS

Aliyun OSS 封装保持职责清晰：

- Manager 管理客户端注册
- Client 负责底层 OSS 客户端构建
- Bucket 负责对象操作和 tracing

## 示例配置

本项目提供了调试配置示例：

- `configs/debug/config.yaml`

典型配置结构：

```yaml
mysql:
  default:
    dsn: "root:root@tcp(127.0.0.1:3306)/micro-kit?charset=utf8mb4&parseTime=True&loc=Local"

redis:
  default:
    addr: "127.0.0.1:6379"

nacos:
  config_center:
    servers:
      - ip_addr: "127.0.0.1"
        port: 8848

server:
  http:
    port: 8080
    enable_nacos: false
```

## 设计约束

- 优先通过配置扫描读取，而不是在代码里写死环境参数
- 默认情况下关闭自动注册，避免“隐式副作用”
- Nacos 相关配置尽量统一到 `protos` 中，便于 YAML / protobuf 解析
- 组件尽量保持“薄封装”，适合作为基础库被业务侧复用
- 多个资源的生命周期要做好 shutdown 清理

## 适用场景

- 微服务基础设施统一封装
- Kratos HTTP / gRPC 服务开发
- 数据库、缓存、消息队列统一接入
- 配置中心接入与服务发现
- 统一 tracing / metrics / logging

## 备注

该项目适合作为微服务基础封装库使用，也可作为业务项目的依赖层，按需引入日志、配置、链路追踪、消息中间件和服务治理能力。