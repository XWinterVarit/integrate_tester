package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/XWinterVarit/integrate_tester/db_viewer/backend/internal/model"
	"github.com/XWinterVarit/integrate_tester/db_viewer/backend/internal/service"
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

func (h *QueryHandler) GetFilters(w http.ResponseWriter, r *http.Request) {
	client := r.PathValue("client")
	table := r.PathValue("table")
	writeJSON(w, h.svc.GetFilters(client, table))
}

func (h *QueryHandler) GetPresetQueries(w http.ResponseWriter, r *http.Request) {
	client := r.PathValue("client")
	table := r.PathValue("table")
	writeJSON(w, h.svc.GetPresetQueries(client, table))
}

func (h *QueryHandler) ResolvePresetQuery(w http.ResponseWriter, r *http.Request) {
	client := r.PathValue("client")
	table := r.PathValue("table")
	indexStr := r.PathValue("index")

	idx, err := strconv.Atoi(indexStr)
	if err != nil {
		writeError(w, "invalid index", http.StatusBadRequest)
		return
	}

	resolved, err := h.svc.ResolvePresetQuery(client, table, idx)
	if err != nil {
		writeError(w, err.Error(), http.StatusBadRequest)
		return
	}
	writeJSON(w, model.ResolvedQueryResponse{ResolvedQuery: resolved})
}
