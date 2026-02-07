package telemetry

import (
	"context"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"

	"github.com/maximhq/bifrost/core/schemas"
)

type telemetryTestLogger struct{}

func (l *telemetryTestLogger) Debug(msg string, args ...any)                     {}
func (l *telemetryTestLogger) Info(msg string, args ...any)                      {}
func (l *telemetryTestLogger) Warn(msg string, args ...any)                      {}
func (l *telemetryTestLogger) Error(msg string, args ...any)                     {}
func (l *telemetryTestLogger) Fatal(msg string, args ...any)                     {}
func (l *telemetryTestLogger) SetLevel(level schemas.LogLevel)                   {}
func (l *telemetryTestLogger) SetOutputType(outputType schemas.LoggerOutputType) {}

func TestRoutingProfileMetricsEmission(t *testing.T) {
	registry := prometheus.NewRegistry()
	pl, err := Init(&Config{Registry: registry}, nil, &telemetryTestLogger{})
	if err != nil {
		t.Fatalf("failed to init telemetry plugin: %v", err)
	}

	ctx := schemas.NewBifrostContext(context.Background(), time.Now().Add(30*time.Second))
	ctx.SetValue(schemas.BifrostContextKey("bf-governance-routing-profile-id"), "rp-light")
	ctx.SetValue(schemas.BifrostContextKey("bf-governance-routing-profile-name"), "Light")
	ctx.SetValue(schemas.BifrostContextKey("bf-governance-routing-profile-fallback-count"), 2)
	ctx.SetValue(schemas.BifrostContextKey("bf-governance-routing-profile-rejections"), map[string]int{"capability_mismatch": 3})

	_, _, _ = pl.PreLLMHook(ctx, nil)

	result := &schemas.BifrostResponse{
		ChatResponse: &schemas.BifrostChatResponse{
			ExtraFields: schemas.BifrostResponseExtraFields{
				RequestType:    schemas.RequestType("chat"),
				Provider:       schemas.ModelProvider("anthropic"),
				ModelRequested: "claude-3-5-haiku-latest",
			},
		},
	}

	_, _, _ = pl.PostLLMHook(ctx, result, nil)

	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		hits := testutil.ToFloat64(pl.RoutingProfileHitsTotal.WithLabelValues("rp-light", "Light"))
		if hits >= 1 {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}

	hits := testutil.ToFloat64(pl.RoutingProfileHitsTotal.WithLabelValues("rp-light", "Light"))
	fallbacks := testutil.ToFloat64(pl.RoutingProfileFallbacksTotal.WithLabelValues("rp-light", "Light"))
	rejections := testutil.ToFloat64(pl.RoutingProfileRejectionsTotal.WithLabelValues("rp-light", "capability_mismatch"))

	if hits < 1 {
		t.Fatalf("expected routing profile hit counter >=1, got %f", hits)
	}
	if fallbacks < 2 {
		t.Fatalf("expected routing profile fallback counter >=2, got %f", fallbacks)
	}
	if rejections < 3 {
		t.Fatalf("expected routing profile rejection counter >=3, got %f", rejections)
	}
}
