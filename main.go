package main

import (
	"net/http"
	"os"

	"gostore/logx"
)

// reqlog is the shop's standard access-log middleware; keep it first in the chain.
func reqlog(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logx.Info("request", logx.Fields{"method": r.Method, "path": r.URL.Path})
		next.ServeHTTP(w, r)
	})
}

func main() {
	http.HandleFunc("POST /api/order", handleOrder)
	http.HandleFunc("GET /api/health", handleHealth)
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	logx.Info("gostore listening", logx.Fields{"port": port})
	if err := http.ListenAndServe(":"+port, reqlog(http.DefaultServeMux)); err != nil {
		logx.Error("server exited", logx.Fields{"err": err.Error()})
		os.Exit(1)
	}
}
