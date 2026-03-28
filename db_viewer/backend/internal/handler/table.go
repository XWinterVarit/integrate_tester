package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/XWinterVarit/integrate_tester/db_viewer/backend/internal/model"
	"github.com/XWinterVarit/integrate_tester/db_viewer/backend/internal/service"
)

type TableHandler struct {
	svc *service.TableService
}

func NewTableHandler(svc *service.TableService) *TableHandler {
	return &TableHandler{svc: svc}
}

func (h *TableHandler) GetRows(w http.ResponseWriter, r *http.Request) {
	client := r.PathValue("client")
	table := r.PathValue("table")

	params := model.RowQueryParams{
		Select:  r.URL.Query().Get("select"),
		Sort:    r.URL.Query().Get("sort"),
		SortDir: r.URL.Query().Get("sort_dir"),
		Limit:   parseLimit(r.URL.Query().Get("limit"), 100),
	}

	rows, err := h.svc.GetRows(r.Context(), client, table, params)
	if err != nil {
		writeError(w, fmt.Sprintf("query error: %v", err), http.StatusInternalServerError)
		return
	}
	writeJSON(w, rows)
}

func (h *TableHandler) GetColumns(w http.ResponseWriter, r *http.Request) {
	client := r.PathValue("client")
	table := r.PathValue("table")

	cols, err := h.svc.GetColumns(r.Context(), client, table)
	if err != nil {
		writeError(w, fmt.Sprintf("query error: %v", err), http.StatusInternalServerError)
		return
	}
	writeJSON(w, cols)
}

func (h *TableHandler) GetConstraints(w http.ResponseWriter, r *http.Request) {
	client := r.PathValue("client")
	table := r.PathValue("table")

	data, err := h.svc.GetConstraints(r.Context(), client, table)
	if err != nil {
		writeError(w, fmt.Sprintf("query error: %v", err), http.StatusInternalServerError)
		return
	}
	writeJSON(w, data)
}

func (h *TableHandler) GetIndexes(w http.ResponseWriter, r *http.Request) {
	client := r.PathValue("client")
	table := r.PathValue("table")

	data, err := h.svc.GetIndexes(r.Context(), client, table)
	if err != nil {
		writeError(w, fmt.Sprintf("query error: %v", err), http.StatusInternalServerError)
		return
	}
	writeJSON(w, data)
}

func (h *TableHandler) GetSize(w http.ResponseWriter, r *http.Request) {
	client := r.PathValue("client")
	table := r.PathValue("table")

	data, err := h.svc.GetTableSize(r.Context(), client, table)
	if err != nil {
		writeError(w, fmt.Sprintf("query error: %v", err), http.StatusInternalServerError)
		return
	}
	writeJSON(w, data)
}

func (h *TableHandler) UpdateCell(w http.ResponseWriter, r *http.Request) {
	client := r.PathValue("client")
	table := r.PathValue("table")

	var req model.UpdateCellRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, "invalid body", http.StatusBadRequest)
		return
	}

	if err := h.svc.UpdateCell(r.Context(), client, table, req); err != nil {
		writeError(w, fmt.Sprintf("update error: %v", err), http.StatusInternalServerError)
		return
	}
	writeJSON(w, model.StatusResponse{Status: "ok"})
}

func (h *TableHandler) DownloadBlob(w http.ResponseWriter, r *http.Request) {
	client := r.PathValue("client")
	table := r.PathValue("table")

	column := r.URL.Query().Get("column")
	whereCol := r.URL.Query().Get("where_column")
	whereVal := r.URL.Query().Get("where_value")

	if column == "" || whereCol == "" || whereVal == "" {
		writeError(w, "missing column, where_column, or where_value", http.StatusBadRequest)
		return
	}

	data, err := h.svc.GetBlobData(r.Context(), client, table, column, whereCol, whereVal)
	if err != nil {
		writeError(w, fmt.Sprintf("blob error: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s_%s.bin"`, table, column))
	w.Write(data)
}

func parseLimit(s string, defaultVal int) int {
	if s == "" {
		return defaultVal
	}
	v, err := strconv.Atoi(s)
	if err != nil || v <= 0 {
		return defaultVal
	}
	return v
}
