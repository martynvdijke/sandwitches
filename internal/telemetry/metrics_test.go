package telemetry

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel/attribute"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
)

func TestMetricsMiddlewareDisabledWithoutInstruments(t *testing.T) {
	if m := (&Telemetry{}).MetricsMiddleware(); m != nil {
		t.Fatal("metrics middleware should be nil without instruments")
	}
}

func TestMetricsMiddlewareRecordsRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)

	reader := sdkmetric.NewManualReader()
	mp := sdkmetric.NewMeterProvider(sdkmetric.WithReader(reader))
	hm, err := newHTTPMetrics(mp.Meter(instrumentationName))
	if err != nil {
		t.Fatalf("newHTTPMetrics: %v", err)
	}
	tel := &Telemetry{Enabled: true, metrics: hm}

	router := gin.New()
	router.Use(tel.MetricsMiddleware())
	router.GET("/hello", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/hello", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}

	var rm metricdata.ResourceMetrics
	if err := reader.Collect(context.Background(), &rm); err != nil {
		t.Fatalf("collect: %v", err)
	}

	points := histogramPoints(t, rm, "http.server.request.duration")
	if len(points) != 1 {
		t.Fatalf("duration data points = %d, want 1", len(points))
	}
	if points[0].Count != 1 {
		t.Errorf("duration count = %d, want 1", points[0].Count)
	}

	assertStringAttr(t, points[0], "http.request.method", http.MethodGet)
	assertStringAttr(t, points[0], "http.route", "/hello")
	assertIntAttr(t, points[0], "http.response.status_code", http.StatusOK)
}

func TestMetricsMiddlewareRouteForUnmatched(t *testing.T) {
	gin.SetMode(gin.TestMode)

	reader := sdkmetric.NewManualReader()
	mp := sdkmetric.NewMeterProvider(sdkmetric.WithReader(reader))
	hm, err := newHTTPMetrics(mp.Meter(instrumentationName))
	if err != nil {
		t.Fatalf("newHTTPMetrics: %v", err)
	}
	tel := &Telemetry{Enabled: true, metrics: hm}

	router := gin.New()
	router.Use(tel.MetricsMiddleware())

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/missing", nil))

	var rm metricdata.ResourceMetrics
	if err := reader.Collect(context.Background(), &rm); err != nil {
		t.Fatalf("collect: %v", err)
	}
	points := histogramPoints(t, rm, "http.server.request.duration")
	if len(points) != 1 {
		t.Fatalf("duration data points = %d, want 1", len(points))
	}
	assertStringAttr(t, points[0], "http.route", "unmatched")
}

func histogramPoints(t *testing.T, rm metricdata.ResourceMetrics, name string) []metricdata.HistogramDataPoint[float64] {
	t.Helper()
	for _, scope := range rm.ScopeMetrics {
		for _, m := range scope.Metrics {
			if m.Name != name {
				continue
			}
			h, ok := m.Data.(metricdata.Histogram[float64])
			if !ok {
				t.Fatalf("metric %s is %T, want histogram", name, m.Data)
			}
			return h.DataPoints
		}
	}
	t.Fatalf("metric %s not found", name)
	return nil
}

func assertStringAttr(t *testing.T, dp metricdata.HistogramDataPoint[float64], key, want string) {
	t.Helper()
	v, ok := dp.Attributes.Value(attribute.Key(key))
	if !ok {
		t.Fatalf("attribute %s missing", key)
	}
	if got := v.AsString(); got != want {
		t.Errorf("attribute %s = %q, want %q", key, got, want)
	}
}

func assertIntAttr(t *testing.T, dp metricdata.HistogramDataPoint[float64], key string, want int64) {
	t.Helper()
	v, ok := dp.Attributes.Value(attribute.Key(key))
	if !ok {
		t.Fatalf("attribute %s missing", key)
	}
	if got := v.AsInt64(); got != want {
		t.Errorf("attribute %s = %d, want %d", key, got, want)
	}
}
