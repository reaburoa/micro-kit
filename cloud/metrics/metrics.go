package metrics

import (
	"github.com/welltop-cn/common/utils/log"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	otelPrometheus "go.opentelemetry.io/otel/exporters/prometheus"
	api "go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
)

var Meter api.Meter

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
	Meter = otel.GetMeterProvider().Meter(serviceName)
}
