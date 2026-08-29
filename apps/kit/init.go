package kit

import (
	"github.com/reaburoa/micro-kit/cloud/config"
	"github.com/reaburoa/micro-kit/utils/env"
	"github.com/reaburoa/micro-kit/utils/log"
)

func Init(serviceName string, ops ...KitOptions) error {
	kitOps := newKitOptions(serviceName)
	env.SetServiceName(kitOps.serviceName)

	for _, o := range ops {
		o(kitOps)
	}

	if err := config.InitConfig(); err != nil {
		return err
	}

	if err := kitOps.runHooks(); err != nil {
		log.Warnf("run init hooks failed: %v", err)
	}

	// 监听退出信号
	go kitOps.waitingShutdown()
	return nil
}
