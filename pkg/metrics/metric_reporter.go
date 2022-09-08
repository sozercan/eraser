package metrics

import (
	"fmt"

	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/metric/instrument"
	"go.opentelemetry.io/otel/metric/instrument/asyncfloat64"
	"go.opentelemetry.io/otel/metric/instrument/syncfloat64"
	"go.opentelemetry.io/otel/metric/unit"
)

var (
	ImageJobCollectorDuration syncfloat64.Histogram
	ImageJobEraserDuration    syncfloat64.Histogram
	PodsRunning               asyncfloat64.Gauge
	ImagesRemoved             syncfloat64.Counter
	VulnerableImages          syncfloat64.Counter
	ImageJobCollectorTotal    syncfloat64.Counter
	ImageJobEraserTotal       syncfloat64.Counter
	PodsCompleted             syncfloat64.Counter
	PodsFailed                syncfloat64.Counter
)

func InitMetricInstruments() {
	meter := metric.NewNoopMeterProvider().Meter("eraser")
	var err error

	if ImageJobCollectorDuration, err = meter.SyncFloat64().Histogram("imagejob_collector_duration", instrument.WithDescription("Distribution of how long it took for collector imagejobs"), instrument.WithUnit(unit.Milliseconds)); err != nil {
		fmt.Println("Failed to register instrument: ImageJobCollectorDuration")
		panic(err)
	}

	if ImageJobEraserDuration, err = meter.SyncFloat64().Histogram("imagejob_eraser_duration", instrument.WithDescription("Distribution of how long it took for eraser imagejobs"), instrument.WithUnit(unit.Milliseconds)); err != nil {
		fmt.Println("Failed to register instrument: ImageJobEraserDuration")
		panic(err)
	}

	if PodsRunning, err = meter.AsyncFloat64().Gauge("pods_running", instrument.WithDescription("Count of total number of collector/eraser pods running"), instrument.WithUnit(unit.Milliseconds)); err != nil {
		fmt.Println("Failed to register instrument: PodsRunning")
		panic(err)
	}

	if ImagesRemoved, err = meter.SyncFloat64().Counter("images_removed", instrument.WithDescription("Count of total number of images removed")); err != nil {
		fmt.Println("Failed to register instrument: ImagesRemoved")
		panic(err)
	}

	if VulnerableImages, err = meter.SyncFloat64().Counter("vulnerable_images", instrument.WithDescription("Count of total number of vulnerable images found")); err != nil {
		fmt.Println("Failed to register instrument: VulnerableImages")
		panic(err)
	}

	if ImageJobCollectorTotal, err = meter.SyncFloat64().Counter("imagejob_collector_total", instrument.WithDescription("Count of total number of collector imagejobs scheduled")); err != nil {
		fmt.Println("Failed to register instrument: ImageJobCollectorTotal")
		panic(err)
	}

	if ImageJobEraserTotal, err = meter.SyncFloat64().Counter("imagejob_eraser_total", instrument.WithDescription("Count of total number of eraser imagejobs scheduled")); err != nil {
		fmt.Println("Failed to register instrument: ImageJobEraserTotal")
		panic(err)
	}

	if PodsCompleted, err = meter.SyncFloat64().Counter("pods_completed", instrument.WithDescription("Count of total number of eraser imagejobs scheduled")); err != nil {
		fmt.Println("Failed to register instrument: PodsCompleted")
		panic(err)
	}

	if PodsFailed, err = meter.SyncFloat64().Counter("pods_failed", instrument.WithDescription("Count of total number of eraser imagejobs scheduled")); err != nil {
		fmt.Println("Failed to register instrument: PodsFailed")
		panic(err)
	}
}
