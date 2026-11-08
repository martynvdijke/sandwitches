package telemetry

import (
	"time"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

// instrumentationName identifies the meters and loggers created by this
// package.
const instrumentationName = "github.com/martynvdijke/sandwitches-go/internal/telemetry"

// httpMetrics holds the instruments used by the request middleware.
type httpMetrics struct {
	active   metric.Int64UpDownCounter
	duration metric.Float64Histogram
}

// newHTTPMetrics creates the HTTP server instruments following the OpenTelemetry
// semantic conventions for HTTP servers.
func newHTTPMetrics(meter metric.Meter) (*httpMetrics, error) {
	active, err := meter.Int64UpDownCounter(
		"http.server.active_requests",
		metric.WithDescription("Number of HTTP requests currently being served."),
		metric.WithUnit("{request}"),
	)
	if err != nil {
		return nil, err
	}

	duration, err := meter.Float64Histogram(
		"http.server.request.duration",
		metric.WithDescription("Duration of HTTP server requests."),
		metric.WithUnit("s"),
		// Seconds-scale buckets: the SDK default targets milliseconds.
		metric.WithExplicitBucketBoundaries(
			0.005, 0.01, 0.025, 0.05, 0.075, 0.1,
			0.25, 0.5, 0.75, 1, 2.5, 5, 7.5, 10,
		),
	)
	if err != nil {
		return nil, err
	}

	return &httpMetrics{active: active, duration: duration}, nil
}

// MetricsMiddleware records HTTP server metrics for every request. It is
// installed after the tracing middleware so the recorded context carries the
// active span. It returns nil when metrics are not configured.
func (t *Telemetry) MetricsMiddleware() gin.HandlerFunc {
	if t == nil || t.metrics == nil {
		return nil
	}

	return func(c *gin.Context) {
		ctx := c.Request.Context()
		start := time.Now()

		t.metrics.active.Add(ctx, 1)
		defer t.metrics.active.Add(ctx, -1)

		c.Next()

		route := c.FullPath()
		if route == "" {
			route = "unmatched"
		}
		t.metrics.duration.Record(ctx, time.Since(start).Seconds(),
			metric.WithAttributes(
				attribute.String("http.request.method", c.Request.Method),
				attribute.String("http.route", route),
				attribute.Int("http.response.status_code", c.Writer.Status()),
			),
		)
	}
}
