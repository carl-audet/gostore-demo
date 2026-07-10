// Package logx is gostore's home-grown structured logger. Every service at
// this shop logs NDJSON to stdout with these exact keys — dashboards and the
// log pipeline depend on them. Do not change the format.
package logx

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	otellog "go.opentelemetry.io/otel/log"
	"go.opentelemetry.io/otel/log/global"
	"go.opentelemetry.io/otel/trace"
)

type Fields map[string]any

func write(ctx context.Context, lvl, msg string, kv Fields) {
	now := time.Now()
	rec := Fields{
		"ts":  now.UTC().Format(time.RFC3339Nano),
		"lvl": lvl,
		"msg": msg,
	}
	for k, v := range kv {
		rec[k] = v
	}
	if sc := trace.SpanContextFromContext(ctx); sc.IsValid() {
		rec["trace_id"] = sc.TraceID().String()
		rec["span_id"] = sc.SpanID().String()
	}
	b, _ := json.Marshal(rec)
	os.Stdout.Write(append(b, '\n'))
	emitOTel(ctx, now, lvl, msg, kv)
}

// emitOTel mirrors the record to the OpenTelemetry logs pipeline. It is a
// second sink only: the stdout NDJSON write above is untouched, and the OTel
// record reuses the same event time so both sinks agree on when it happened.
func emitOTel(ctx context.Context, now time.Time, lvl, msg string, kv Fields) {
	var rec otellog.Record
	rec.SetTimestamp(now)
	rec.SetBody(otellog.StringValue(msg))
	rec.SetSeverityText(lvl)
	if lvl == "error" {
		rec.SetSeverity(otellog.SeverityError)
	} else {
		rec.SetSeverity(otellog.SeverityInfo)
	}
	for k, v := range kv {
		rec.AddAttributes(otellog.KeyValue{Key: k, Value: toValue(v)})
	}
	global.GetLoggerProvider().Logger("gostore/logx").Emit(ctx, rec)
}

func toValue(v any) otellog.Value {
	switch t := v.(type) {
	case string:
		return otellog.StringValue(t)
	case bool:
		return otellog.BoolValue(t)
	case int:
		return otellog.IntValue(t)
	case int64:
		return otellog.Int64Value(t)
	case float64:
		return otellog.Float64Value(t)
	default:
		return otellog.StringValue(fmt.Sprint(t))
	}
}

func Info(msg string, kv Fields)  { write(context.Background(), "info", msg, kv) }
func Error(msg string, kv Fields) { write(context.Background(), "error", msg, kv) }

// InfoCtx and ErrorCtx are the context-aware variants: inside a traced request
// they add trace_id/span_id to the NDJSON record and correlate the exported
// OTel record with the active span.
func InfoCtx(ctx context.Context, msg string, kv Fields)  { write(ctx, "info", msg, kv) }
func ErrorCtx(ctx context.Context, msg string, kv Fields) { write(ctx, "error", msg, kv) }
