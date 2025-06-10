package kit

import (
	"github.com/reaburoa/micro-kit/cloud/config"
	"github.com/reaburoa/micro-kit/utils/env"
	"github.com/reaburoa/micro-kit/utils/log"
)

func Init(serviceName string, ops ...KitOptions) error {
	kitOps := &kitOptions{
		serviceName: serviceName,
	}
	env.SetServiceName(kitOps.serviceName)

	config.InitConfig()

	log.InitLogger()

	//	metrics.InitMetrics(kitOps.serviceName)

	for _, o := range ops {
		o(kitOps)
	}
	// 监听退出信号
	go kitOps.waitingShutdown()
	return nil
}
