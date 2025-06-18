package kratos

import (
	"github.com/go-kratos/kratos/v2/middleware/metrics"
	kitMetrics "github.com/reaburoa/micro-kit/cloud/metrics"
	"github.com/reaburoa/micro-kit/utils/log"
	"go.opentelemetry.io/otel/metric"
)

func MetricsRequests() metric.Int64Counter {
	metricRequests, err := metrics.DefaultRequestsCounter(kitMetrics.Meter, metrics.DefaultServerRequestsCounterName)
	if err != nil {
		log.Fatalf("init kratos request metrics failed with err %s", err.Error())
	}

	return metricRequests
}

func MetricsSeconds() metric.Float64Histogram {
	metricSeconds, err := metrics.DefaultSecondsHistogram(kitMetrics.Meter, metrics.DefaultServerSecondsHistogramName)
	if err != nil {
		log.Fatalf("init kratos seconds metrics failed with err %s", err.Error())
	}

	return metricSeconds
}
