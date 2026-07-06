// Package logx is gostore's home-grown structured logger. Every service at
// this shop logs NDJSON to stdout with these exact keys — dashboards and the
// log pipeline depend on them. Do not change the format.
package logx

import (
	"encoding/json"
	"os"
	"time"
)

type Fields map[string]any

func write(lvl, msg string, kv Fields) {
	rec := Fields{
		"ts":  time.Now().UTC().Format(time.RFC3339Nano),
		"lvl": lvl,
		"msg": msg,
	}
	for k, v := range kv {
		rec[k] = v
	}
	b, _ := json.Marshal(rec)
	os.Stdout.Write(append(b, '\n'))
}

func Info(msg string, kv Fields)  { write("info", msg, kv) }
func Error(msg string, kv Fields) { write("error", msg, kv) }
