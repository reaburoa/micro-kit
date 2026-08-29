package metrics

import (
	"sync"

	"github.com/reaburoa/micro-kit/utils/log"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	otelPrometheus "go.opentelemetry.io/otel/exporters/prometheus"
	api "go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
)

var (
	meterMu sync.RWMutex
	Meter   api.Meter = otel.GetMeterProvider().Meter("micro-kit-default")
)

func GetMeter() api.Meter {
	meterMu.RLock()
	defer meterMu.RUnlock()
	if Meter == nil {
		return otel.GetMeterProvider().Meter("micro-kit-default")
	}
	return Meter
}

func InitMetrics(serviceName string) {
	// The exporter embeds a default OpenTelemetry Reader and
	// implements prometheus.Collector, allowing it to be used as
	// both a Reader and Collector.
	exporter, err := otelPrometheus.New(otelPrometheus.WithNamespace(serviceName), otelPrometheus.WithoutTargetInfo())
	if err != nil {
		log.Fatal("init otelPrometheus exporter failed with ", err)
	}

	res, err := resource.Merge(resource.Default(),
		resource.NewWithAttributes(resource.Default().SchemaURL(), attribute.Key("service.name").String(serviceName)),
	)
	if err != nil {
		log.Fatal("init exporter resource failed with ", err)
	}
	provider := metric.NewMeterProvider(metric.WithResource(res), metric.WithReader(exporter))
	otel.SetMeterProvider(provider)
	meterMu.Lock()
	Meter = provider.Meter(serviceName)
	meterMu.Unlock()
}
