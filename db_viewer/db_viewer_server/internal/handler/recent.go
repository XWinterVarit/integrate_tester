package handler

import (
	"encoding/json"
	"net/http"

	"github.com/XWinterVarit/integrate_tester/db_viewer/db_viewer_server/internal/model"
	"github.com/XWinterVarit/integrate_tester/db_viewer/db_viewer_server/internal/service"
)

type RecentHandler struct {
	svc *service.TableService
}

func NewRecentHandler(svc *service.TableService) *RecentHandler {
	return &RecentHandler{svc: svc}
}

func (h *RecentHandler) TouchFilter(w http.ResponseWriter, r *http.Request) {
	var req model.RecentTouchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, "invalid body", http.StatusBadRequest)
		return
	}
	h.svc.TouchRecentFilter(req.Key)
	writeJSON(w, model.StatusResponse{Status: "ok"})
}

func (h *RecentHandler) TouchQuery(w http.ResponseWriter, r *http.Request) {
	var req model.RecentTouchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, "invalid body", http.StatusBadRequest)
		return
	}
	h.svc.TouchRecentQuery(req.Key)
	writeJSON(w, model.StatusResponse{Status: "ok"})
}
