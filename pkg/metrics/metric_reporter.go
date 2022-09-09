package metrics

import (
	"fmt"
	"net/http"
	"os"

	runtimemetrics "go.opentelemetry.io/contrib/instrumentation/runtime"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/metric/instrument"
	"go.opentelemetry.io/otel/metric/instrument/asyncfloat64"
	"go.opentelemetry.io/otel/metric/instrument/syncfloat64"
	"go.opentelemetry.io/otel/metric/unit"
	"k8s.io/klog/v2"
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
	metricsAddr               = ":8088"
)

func InitMetricInstruments() error {
	err, exporter := InitPrometheusExporter(metricsAddr)
	if err != nil {
		fmt.Println("unable to initialize prometheus exporter")
		return err
	}

	if err := runtimemetrics.Start(); err != nil {
		fmt.Println("unable to start metrics")
		return err
	}

	http.HandleFunc("/metrics", exporter.ServeHTTP)
	go func() {
		if err := http.ListenAndServe(metricsAddr, nil); err != nil {
			klog.ErrorS(err, "failed to register prometheus endpoint", "metricsAddress", metricsAddr)
			os.Exit(1)
		}
	}()

	klog.InfoS("Prometheus metrics server running", "address", metricsAddr)

	meter := metric.NewNoopMeterProvider().Meter("eraser")

	if ImageJobCollectorDuration, err = meter.SyncFloat64().Histogram("imagejob_collector_duration", instrument.WithDescription("Distribution of how long it took for collector imagejobs"), instrument.WithUnit(unit.Milliseconds)); err != nil {
		klog.InfoS("Failed to register instrument: ImageJobCollectorDuration")
		return err
	}

	if ImageJobEraserDuration, err = meter.SyncFloat64().Histogram("imagejob_eraser_duration", instrument.WithDescription("Distribution of how long it took for eraser imagejobs"), instrument.WithUnit(unit.Milliseconds)); err != nil {
		klog.InfoS("Failed to register instrument: ImageJobEraserDuration")
		return err
	}

	/*
		if PodsRunning, err = meter.AsyncFloat64().Gauge("pods_running", instrument.WithDescription("Count of total number of collector/eraser pods running"), instrument.WithUnit(unit.Milliseconds)); err != nil {
			klog.InfoS("Failed to register instrument: PodsRunning")
			return err
		}

		if ImagesRemoved, err = meter.SyncFloat64().Counter("images_removed", instrument.WithDescription("Count of total number of images removed")); err != nil {
			klog.InfoS("Failed to register instrument: ImagesRemoved")
			return err
		}

		if VulnerableImages, err = meter.SyncFloat64().Counter("vulnerable_images", instrument.WithDescription("Count of total number of vulnerable images found")); err != nil {
			klog.InfoS("Failed to register instrument: VulnerableImages")
			return err
		}

		if ImageJobCollectorTotal, err = meter.SyncFloat64().Counter("imagejob_collector_total", instrument.WithDescription("Count of total number of collector imagejobs scheduled")); err != nil {
			klog.InfoS("Failed to register instrument: ImageJobCollectorTotal")
			return err
		}

		if ImageJobEraserTotal, err = meter.SyncFloat64().Counter("imagejob_eraser_total", instrument.WithDescription("Count of total number of eraser imagejobs scheduled")); err != nil {
			klog.InfoS("Failed to register instrument: ImageJobEraserTotal")
			return err
		}

		if PodsCompleted, err = meter.SyncFloat64().Counter("pods_completed", instrument.WithDescription("Count of total number of eraser imagejobs scheduled")); err != nil {
			klog.InfoS("Failed to register instrument: PodsCompleted")
			return err
		}

		if PodsFailed, err = meter.SyncFloat64().Counter("pods_failed", instrument.WithDescription("Count of total number of eraser imagejobs scheduled")); err != nil {
			klog.InfoS("Failed to register instrument: PodsFailed")
			return err
		}
	*/
	return nil
}
