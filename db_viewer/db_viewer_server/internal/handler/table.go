package handler

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/XWinterVarit/integrate_tester/db_viewer/db_viewer_server/internal/model"
	"github.com/XWinterVarit/integrate_tester/db_viewer/db_viewer_server/internal/service"
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
		Where:   r.URL.Query().Get("where"),
		Sort:    r.URL.Query().Get("sort"),
		SortDir: r.URL.Query().Get("sort_dir"),
		Limit:   parseLimit(r.URL.Query().Get("limit"), 100),
		Offset:  parseOffset(r.URL.Query().Get("offset")),
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

func (h *TableHandler) GetRowCount(w http.ResponseWriter, r *http.Request) {
	client := r.PathValue("client")
	table := r.PathValue("table")

	count, err := h.svc.GetRowCount(r.Context(), client, table)
	if err != nil {
		writeError(w, fmt.Sprintf("query error: %v", err), http.StatusInternalServerError)
		return
	}
	writeJSON(w, map[string]int{"count": count})
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

func (h *TableHandler) DeleteRow(w http.ResponseWriter, r *http.Request) {
	client := r.PathValue("client")
	table := r.PathValue("table")

	var req model.DeleteRowRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, "invalid body", http.StatusBadRequest)
		return
	}

	if err := h.svc.DeleteRow(r.Context(), client, table, req); err != nil {
		writeError(w, fmt.Sprintf("delete error: %v", err), http.StatusInternalServerError)
		return
	}
	writeJSON(w, model.StatusResponse{Status: "ok"})
}

func (h *TableHandler) InsertRow(w http.ResponseWriter, r *http.Request) {
	client := r.PathValue("client")
	table := r.PathValue("table")

	var req model.InsertRowRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, "invalid body", http.StatusBadRequest)
		return
	}

	rowid, err := h.svc.InsertRow(r.Context(), client, table, req)
	if err != nil {
		writeError(w, fmt.Sprintf("insert error: %v", err), http.StatusInternalServerError)
		return
	}
	writeJSON(w, map[string]string{"status": "ok", "rowid": rowid})
}

func (h *TableHandler) BuildDeleteQuery(w http.ResponseWriter, r *http.Request) {
	client := r.PathValue("client")
	table := r.PathValue("table")
	rowid := r.URL.Query().Get("rowid")

	if rowid == "" {
		writeError(w, "missing rowid", http.StatusBadRequest)
		return
	}

	query, err := h.svc.BuildDeleteQuery(client, table, rowid)
	if err != nil {
		writeError(w, fmt.Sprintf("error: %v", err), http.StatusInternalServerError)
		return
	}
	writeJSON(w, map[string]string{"query": query})
}

func (h *TableHandler) BuildUpdateQuery(w http.ResponseWriter, r *http.Request) {
	client := r.PathValue("client")
	table := r.PathValue("table")
	column := r.URL.Query().Get("column")
	value := r.URL.Query().Get("value")
	rowid := r.URL.Query().Get("rowid")

	if column == "" || rowid == "" {
		writeError(w, "missing column or rowid", http.StatusBadRequest)
		return
	}

	query, err := h.svc.BuildUpdateQuery(client, table, column, value, rowid)
	if err != nil {
		writeError(w, fmt.Sprintf("error: %v", err), http.StatusInternalServerError)
		return
	}
	writeJSON(w, map[string]string{"query": query})
}

func (h *TableHandler) BuildInsertQuery(w http.ResponseWriter, r *http.Request) {
	client := r.PathValue("client")
	table := r.PathValue("table")

	var req model.InsertRowRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, "invalid body", http.StatusBadRequest)
		return
	}

	query, err := h.svc.BuildInsertQuery(client, table, req.Columns, req.Values)
	if err != nil {
		writeError(w, fmt.Sprintf("error: %v", err), http.StatusInternalServerError)
		return
	}
	writeJSON(w, map[string]string{"query": query})
}

func (h *TableHandler) UploadBlob(w http.ResponseWriter, r *http.Request) {
	client := r.PathValue("client")
	table := r.PathValue("table")
	column := r.URL.Query().Get("column")
	rowid := r.URL.Query().Get("rowid")
	if column == "" || rowid == "" {
		writeError(w, "missing column or rowid", http.StatusBadRequest)
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, 50<<20) // 50 MB limit
	data, err := io.ReadAll(r.Body)
	if err != nil {
		writeError(w, fmt.Sprintf("read error: %v", err), http.StatusBadRequest)
		return
	}
	if err := h.svc.UploadBlobData(r.Context(), client, table, column, rowid, data); err != nil {
		writeError(w, fmt.Sprintf("upload error: %v", err), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"ok":true}`))
}

func (h *TableHandler) DownloadBlob(w http.ResponseWriter, r *http.Request) {
	client := r.PathValue("client")
	table := r.PathValue("table")

	column := r.URL.Query().Get("column")
	rowid := r.URL.Query().Get("rowid")

	if column == "" || rowid == "" {
		writeError(w, "missing column or rowid", http.StatusBadRequest)
		return
	}

	data, err := h.svc.GetBlobData(r.Context(), client, table, column, rowid)
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

func parseOffset(s string) int {
	if s == "" {
		return 0
	}
	v, err := strconv.Atoi(s)
	if err != nil || v < 0 {
		return 0
	}
	return v
}
