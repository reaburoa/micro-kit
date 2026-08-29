package kit

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/reaburoa/micro-kit/cloud/config"
	"github.com/reaburoa/micro-kit/cloud/metrics"
	"github.com/reaburoa/micro-kit/cloud/tracer"
	"github.com/reaburoa/micro-kit/protos"
	"github.com/reaburoa/micro-kit/utils/log"
)

type initHook func(*kitOptions) error

type kitOptions struct {
	serviceName  string
	shutdownFunc []func(ctx context.Context) error
	nacosConfig  *protos.NacosConfigCenter
	useLocalOnly bool
	initHooks    []initHook
}

func (k *kitOptions) addShutdownFunc(f func(ctx context.Context) error) {
	if k.shutdownFunc == nil {
		k.shutdownFunc = make([]func(ctx context.Context) error, 0, 4)
	}
	k.shutdownFunc = append(k.shutdownFunc, f)
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
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		err := f(ctx)
		cancel()
		if err != nil {
			log.Errorf("shutdown error, %v", err)
		}
	}
}

func (k *kitOptions) addInitHook(h initHook) {
	if k.initHooks == nil {
		k.initHooks = make([]initHook, 0, 4)
	}
	k.initHooks = append(k.initHooks, h)
}

func (k *kitOptions) runHooks() error {
	log.InitLogger()
	metrics.InitMetrics(k.serviceName)

	for _, hook := range k.initHooks {
		if err := hook(k); err != nil {
			log.Warnf("init hook failed: %v", err)
			return err
		}
	}

	return nil
}

type KitOptions func(o *kitOptions)

func newKitOptions(serviceName string) *kitOptions {
	k := &kitOptions{serviceName: serviceName}
	k.addInitHook(func(o *kitOptions) error {
		log.InitLogger()
		return nil
	})
	k.addInitHook(func(o *kitOptions) error {
		metrics.InitMetrics(o.serviceName)
		return nil
	})
	return k
}

func WithTracer() KitOptions {
	return func(o *kitOptions) {
		o.addInitHook(func(k *kitOptions) error {
			log.Infof("==== init otel tracing ===")
			shutdown, err := tracer.InitOtelTracer()
			if err != nil {
				log.Errorf("failed to init otel tracer: %v", err)
				return err
			}
			if shutdown != nil {
				k.addShutdownFunc(shutdown)
				log.Infof("=== init otel tracing success ===")
			}
			return nil
		})
	}
}

func WithNacosConfig(cfg *protos.NacosConfigCenter) KitOptions {
	return func(o *kitOptions) {
		o.nacosConfig = cfg
		o.useLocalOnly = false
		if cfg != nil {
			config.SetNacosConfigOverride(cfg)
			return
		}
		config.SetLocalConfigOnly()
	}
}

func WithLocalConfig() KitOptions {
	return func(o *kitOptions) {
		o.nacosConfig = nil
		o.useLocalOnly = true
		config.SetLocalConfigOnly()
	}
}
