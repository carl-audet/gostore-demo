package main

import (
	"encoding/json"
	"net/http"

	"gostore/logx"
)

type orderReq struct {
	Items []struct {
		SKU string  `json:"sku"`
		Qty int     `json:"qty"`
		EUR float64 `json:"eur"`
	} `json:"items"`
	Email string `json:"email"`
}

func handleOrder(w http.ResponseWriter, r *http.Request) {
	var req orderReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "bad json"})
		return
	}
	if len(req.Items) == 0 {
		w.WriteHeader(http.StatusUnprocessableEntity)
		json.NewEncoder(w).Encode(map[string]string{"error": "no items"})
		return
	}
	total := 0.0
	for _, it := range req.Items {
		total += it.EUR * float64(it.Qty)
	}
	logx.Info("order accepted", logx.Fields{"items": len(req.Items), "total_eur": total})
	json.NewEncoder(w).Encode(map[string]any{"ok": true, "total_eur": total})
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	logx.Info("health check", logx.Fields{"ok": true})
	json.NewEncoder(w).Encode(map[string]string{"status": "ok", "service": "gostore"})
}
