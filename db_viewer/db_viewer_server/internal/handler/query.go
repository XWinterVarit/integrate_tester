package handler

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/XWinterVarit/integrate_tester/db_viewer/db_viewer_server/internal/model"
	"github.com/XWinterVarit/integrate_tester/db_viewer/db_viewer_server/internal/service"
)

type QueryHandler struct {
	svc *service.TableService
}

func NewQueryHandler(svc *service.TableService) *QueryHandler {
	return &QueryHandler{svc: svc}
}

func (h *QueryHandler) Execute(w http.ResponseWriter, r *http.Request) {
	client := r.PathValue("client")

	var req model.ExecuteQueryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, "invalid body", http.StatusBadRequest)
		return
	}

	rows, err := h.svc.ExecuteQuery(r.Context(), client, req)
	if err != nil {
		writeError(w, fmt.Sprintf("query error: %v", err), http.StatusInternalServerError)
		return
	}
	writeJSON(w, rows)
}
