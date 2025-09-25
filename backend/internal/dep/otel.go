package dep

import (
	"fmt"

	"stream-server/config"

	"github.com/go-kratos/kratos/v2/middleware/metrics"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/prometheus"
	"go.opentelemetry.io/otel/metric"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
)

func NewMeter(cfg *conf.Metadata, provider metric.MeterProvider) (metric.Meter, error) {
	name := fmt.Sprintf("%s-%s", cfg.Namespace, cfg.Name)

	return provider.Meter(name), nil
}

func NewMeterProvider(cfg *conf.Config) (metric.MeterProvider, error) {
	meta := cfg.Metadata
	metricConf := cfg.Metrics
	exporter, err := prometheus.New()
	if err != nil {
		return nil, err
	}

	if metricConf.Enabled {
		err = metrics.EnableOTELExemplar()
		if err != nil {
			return nil, err
		}
	}

	provider := sdkmetric.NewMeterProvider(
		sdkmetric.WithResource(
			resource.NewWithAttributes(
				semconv.SchemaURL,
				semconv.ServiceNameKey.String(meta.Name),
				attribute.String("environment", meta.Env),
			),
		),
		sdkmetric.WithReader(exporter),
		sdkmetric.WithView(
			metrics.DefaultSecondsHistogramView(metrics.DefaultServerSecondsHistogramName),
		),
	)
	otel.SetMeterProvider(provider)
	return provider, nil
}
