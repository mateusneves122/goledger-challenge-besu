package simplestorage

import (
	"encoding/json"
	"log/slog"
	"math/big"
	"net/http"
)

const uint256Bits = 256

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) SetHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Value string `json:"value"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	if req.Value == "" {
		writeError(w, http.StatusBadRequest, "value is required")
		return
	}

	n, ok := new(big.Int).SetString(req.Value, 10)
	if !ok || n.Sign() < 0 {
		writeError(w, http.StatusBadRequest, "value must be a non-negative integer")
		return
	}
	if n.BitLen() > uint256Bits {
		writeError(w, http.StatusBadRequest, "value exceeds uint256 maximum")
		return
	}

	txHash, err := h.service.Set(r.Context(), n)
	if err != nil {
		slog.Error("set failed", "err", err)
		writeError(w, http.StatusInternalServerError, "failed to set value")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"tx_hash": txHash,
		"value":   req.Value,
	})
}

func (h *Handler) GetHandler(w http.ResponseWriter, r *http.Request) {
	value, err := h.service.Get(r.Context())
	if err != nil {
		slog.Error("get failed", "err", err)
		writeError(w, http.StatusInternalServerError, "failed to get value")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"value": value})
}

func (h *Handler) SyncHandler(w http.ResponseWriter, r *http.Request) {
	value, err := h.service.Sync(r.Context())
	if err != nil {
		slog.Error("sync failed", "err", err)
		writeError(w, http.StatusInternalServerError, "failed to sync value")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"value":   value,
		"message": "synced",
	})
}

func (h *Handler) CheckHandler(w http.ResponseWriter, r *http.Request) {
	match, err := h.service.Check(r.Context())
	if err != nil {
		slog.Error("check failed", "err", err)
		writeError(w, http.StatusInternalServerError, "failed to check values")
		return
	}

	writeJSON(w, http.StatusOK, map[string]bool{"match": match})
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}
