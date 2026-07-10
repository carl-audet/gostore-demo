package main

import (
	"context"
	"net/http"
	"os"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"

	"gostore/logx"
)

// reqlog is the shop's standard access-log middleware; keep it first in the chain.
// (otelhttp sits outside it only so the access log can carry the trace id.)
func reqlog(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logx.InfoCtx(r.Context(), "request", logx.Fields{"method": r.Method, "path": r.URL.Path})
		next.ServeHTTP(w, r)
	})
}

func main() {
	ctx := context.Background()
	shutdown, err := setupOTel(ctx)
	if err != nil {
		logx.Error("otel init failed", logx.Fields{"err": err.Error()})
		os.Exit(1)
	}
	defer shutdown(ctx)

	http.HandleFunc("POST /api/order", handleOrder)
	http.HandleFunc("GET /api/health", handleHealth)
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	logx.Info("gostore listening", logx.Fields{"port": port})
	if err := http.ListenAndServe(":"+port, otelhttp.NewHandler(reqlog(http.DefaultServeMux), "gostore")); err != nil {
		logx.Error("server exited", logx.Fields{"err": err.Error()})
		os.Exit(1)
	}
}
