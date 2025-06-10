package config

import (
	"fmt"
	"path"

	"github.com/go-kratos/kratos/v2/config"
	"github.com/go-kratos/kratos/v2/config/file"
	"github.com/welltop-cn/common/utils/env"
)

var defaultConfig config.Config

func loadLocalConfig(confPath string) config.Config {
	configPath := path.Join(confPath, fmt.Sprintf("configs/%s", env.GetRuntimeEnv()))
	c := config.New(config.WithSource(file.NewSource(fmt.Sprintf("%s/config.yaml", configPath))))
	defer c.Close() // 关闭watch,不进行自动更新

	if err := c.Load(); err != nil {
		panic(err)
	}

	return c
}

func setConfig(c config.Config) {
	defaultConfig = c
}

func InitConfig() {
	var (
		confPath string
		err      error
	)
	if env.IsDebug() {
		confPath, err = env.GetProjectPath()
		if err != nil {
			panic("get root path " + err.Error())
		}
	}
	conf := loadLocalConfig(confPath)
	setConfig(conf)
}

func Get(key string) config.Value {
	return defaultConfig.Value(key)
}
