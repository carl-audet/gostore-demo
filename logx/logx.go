// Package logx is gostore's home-grown structured logger. Every service at
// this shop logs NDJSON to stdout with these exact keys — dashboards and the
// log pipeline depend on them. Do not change the format.
//
// Each record is additionally emitted through the OpenTelemetry logs API so it
// can be exported over OTLP. The stdout output is untouched; when no OTel
// logger provider is installed the emit is a no-op. Use the Ctx variants
// inside request handlers so exported records carry trace_id/span_id.
package logx

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	otellog "go.opentelemetry.io/otel/log"
	"go.opentelemetry.io/otel/log/global"
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
	b, _ := json.Marshal(rec)
	os.Stdout.Write(append(b, '\n'))
	emitOTel(ctx, now, lvl, msg, kv)
}

// emitOTel mirrors the record to the OTel logs API, reusing the same event
// time as the stdout line.
func emitOTel(ctx context.Context, t time.Time, lvl, msg string, kv Fields) {
	var rec otellog.Record
	rec.SetTimestamp(t)
	rec.SetBody(otellog.StringValue(msg))
	rec.SetSeverityText(lvl)
	if lvl == "error" {
		rec.SetSeverity(otellog.SeverityError)
	} else {
		rec.SetSeverity(otellog.SeverityInfo)
	}
	for k, v := range kv {
		rec.AddAttributes(otellog.KeyValue{Key: k, Value: otelValue(v)})
	}
	global.GetLoggerProvider().Logger("gostore/logx").Emit(ctx, rec)
}

func otelValue(v any) otellog.Value {
	switch v := v.(type) {
	case string:
		return otellog.StringValue(v)
	case bool:
		return otellog.BoolValue(v)
	case int:
		return otellog.IntValue(v)
	case int64:
		return otellog.Int64Value(v)
	case float64:
		return otellog.Float64Value(v)
	default:
		return otellog.StringValue(fmt.Sprint(v))
	}
}

func Info(msg string, kv Fields)  { write(context.Background(), "info", msg, kv) }
func Error(msg string, kv Fields) { write(context.Background(), "error", msg, kv) }

// InfoCtx and ErrorCtx behave exactly like Info/Error but carry the request
// context so the exported log record is correlated with the active span.
func InfoCtx(ctx context.Context, msg string, kv Fields)  { write(ctx, "info", msg, kv) }
func ErrorCtx(ctx context.Context, msg string, kv Fields) { write(ctx, "error", msg, kv) }
