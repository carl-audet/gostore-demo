package main

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

	"gostore/logx"
)

// reqlog is the shop's standard access-log middleware; keep it first in the chain.
func reqlog(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logx.InfoCtx(r.Context(), "request", logx.Fields{"method": r.Method, "path": r.URL.Path})
		next.ServeHTTP(w, r)
	})
}

// routeSpan names the server span after the matched mux pattern (e.g.
// "POST /api/order") and records http.route, which is only known after
// routing. It must wrap the mux directly.
func routeSpan(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r)
		if r.Pattern == "" {
			return
		}
		span := trace.SpanFromContext(r.Context())
		span.SetName(r.Pattern)
		route := r.Pattern
		if i := strings.IndexByte(route, ' '); i >= 0 {
			route = route[i+1:]
		}
		span.SetAttributes(attribute.String("http.route", route))
	})
}

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	otelShutdown, err := setupOTel(ctx)
	if err != nil {
		logx.Error("otel setup failed", logx.Fields{"err": err.Error()})
		os.Exit(1)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/order", handleOrder)
	mux.HandleFunc("GET /api/health", handleHealth)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	srv := &http.Server{
		Addr:    ":" + port,
		Handler: otelhttp.NewHandler(reqlog(routeSpan(mux)), "gostore"),
	}

	logx.Info("gostore listening", logx.Fields{"port": port})
	errCh := make(chan error, 1)
	go func() { errCh <- srv.ListenAndServe() }()

	var srvErr error
	select {
	case srvErr = <-errCh:
	case <-ctx.Done():
		shutCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		srv.Shutdown(shutCtx)
		cancel()
	}

	flushCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	if err := otelShutdown(flushCtx); err != nil {
		logx.Error("otel shutdown failed", logx.Fields{"err": err.Error()})
	}
	cancel()

	if srvErr != nil && !errors.Is(srvErr, http.ErrServerClosed) {
		logx.Error("server exited", logx.Fields{"err": srvErr.Error()})
		os.Exit(1)
	}
}
