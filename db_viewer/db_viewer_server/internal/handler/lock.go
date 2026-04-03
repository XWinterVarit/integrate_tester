package handler

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/XWinterVarit/integrate_tester/db_viewer/db_viewer_server/internal/model"
	"github.com/XWinterVarit/integrate_tester/db_viewer/db_viewer_server/internal/service"
)

type LockHandler struct {
	svc *service.LockService
}

func NewLockHandler(svc *service.LockService) *LockHandler {
	return &LockHandler{svc: svc}
}

func (h *LockHandler) Acquire(w http.ResponseWriter, r *http.Request) {
	var req model.AcquireLockRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, "invalid body", http.StatusBadRequest)
		return
	}
	lock, err := h.svc.Acquire(r.Context(), req)
	if err != nil {
		if strings.HasPrefix(err.Error(), "LOCKED:") {
			writeError(w, strings.TrimPrefix(err.Error(), "LOCKED:"), http.StatusLocked)
			return
		}
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, lock)
}

func (h *LockHandler) Renew(w http.ResponseWriter, r *http.Request) {
	key := r.PathValue("key")
	scopeClient := r.URL.Query().Get("scope_client")
	var req model.RenewLockRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, "invalid body", http.StatusBadRequest)
		return
	}
	lock, err := h.svc.Renew(r.Context(), key, scopeClient, req.SessionID)
	if err != nil {
		if strings.HasPrefix(err.Error(), "LOCKED:") {
			writeError(w, strings.TrimPrefix(err.Error(), "LOCKED:"), http.StatusLocked)
			return
		}
		writeError(w, err.Error(), http.StatusNotFound)
		return
	}
	writeJSON(w, lock)
}

func (h *LockHandler) Release(w http.ResponseWriter, r *http.Request) {
	key := r.PathValue("key")
	scopeClient := r.URL.Query().Get("scope_client")
	var req model.ReleaseLockRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, "invalid body", http.StatusBadRequest)
		return
	}
	if err := h.svc.Release(r.Context(), key, scopeClient, req.SessionID); err != nil {
		if strings.HasPrefix(err.Error(), "LOCKED:") {
			writeError(w, strings.TrimPrefix(err.Error(), "LOCKED:"), http.StatusLocked)
			return
		}
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, model.StatusResponse{Status: "ok"})
}

func (h *LockHandler) GetLock(w http.ResponseWriter, r *http.Request) {
	key := r.PathValue("key")
	scopeClient := r.URL.Query().Get("scope_client")
	lock, err := h.svc.GetLock(r.Context(), key, scopeClient)
	if err != nil {
		writeError(w, err.Error(), http.StatusNotFound)
		return
	}
	writeJSON(w, lock)
}
