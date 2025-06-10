package kit

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/welltop-cn/common/cloud/tracer"
	"github.com/welltop-cn/common/utils/log"
)

type kitOptions struct {
	serviceName  string
	shutdownFunc []func(ctx context.Context) error
}

func (k *kitOptions) waitingShutdown() {
	defer func() {
		if err := recover(); err != nil {
			log.Errorf("panic error, %v", err)
		}
	}()
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	<-signalChan
	log.Infof("receive signal, start to shutdown")
	for index, f := range k.shutdownFunc {
		log.Infof("shutdownFunc index: %d", index)
		err := f(context.Background())
		if err != nil {
			log.Errorf("shutdown error, %v", err)
		}
	}
}

type KitOptions func(o *kitOptions)

func WithTracer() KitOptions {
	return func(o *kitOptions) {
		log.Infof("==== init otel tracing ===")
		shutdown, err := tracer.InitOtelTracer()
		if err != nil {
			log.Errorf("failed to init otel tracer ", err)
		}
		if shutdown != nil {
			if len(o.shutdownFunc) <= 0 {
				o.shutdownFunc = make([]func(ctx context.Context) error, 0, 5)
			}
			o.shutdownFunc = append(o.shutdownFunc, shutdown)
			log.Infof("=== init otel tracing success ===")
		}
	}
}
