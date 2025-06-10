GOHOSTOS:=$(shell go env GOHOSTOS)
GOPATH:=$(shell go env GOPATH)
VERSION=$(shell git describe --tags --always)

PROJECT_PATH:=$(shell pwd)
CONF_PROTO_FILES=$(shell find protos -name *.proto)

.PHONY: config
# 生成配置文件和枚举
config:
	protoc --proto_path=./ --go_out=paths=source_relative:. $(CONF_PROTO_FILES)