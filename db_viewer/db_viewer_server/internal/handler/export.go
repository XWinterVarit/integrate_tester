package handler

import (
	"fmt"
	"net/http"

	"github.com/XWinterVarit/integrate_tester/db_viewer/db_viewer_server/internal/model"
	"github.com/XWinterVarit/integrate_tester/db_viewer/db_viewer_server/internal/service"
)

type ExportHandler struct {
	svc *service.TableService
}

func NewExportHandler(svc *service.TableService) *ExportHandler {
	return &ExportHandler{svc: svc}
}

func (h *ExportHandler) Export(w http.ResponseWriter, r *http.Request) {
	client := r.PathValue("client")
	table := r.PathValue("table")

	exportType := r.URL.Query().Get("type")
	format := r.URL.Query().Get("format")
	if format == "" {
		format = "csv"
	}

	params := model.RowQueryParams{
		Select:  r.URL.Query().Get("select"),
		Sort:    r.URL.Query().Get("sort"),
		SortDir: r.URL.Query().Get("sort_dir"),
		Limit:   parseLimit(r.URL.Query().Get("limit"), 0),
	}

	if format == "json" {
		w.Header().Set("Content-Type", "application/json")
	} else {
		w.Header().Set("Content-Type", "text/csv")
	}
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s.%s", table, format))

	if err := h.svc.ExportTable(r.Context(), w, client, table, exportType, format, params); err != nil {
		writeError(w, fmt.Sprintf("export error: %v", err), http.StatusInternalServerError)
	}
}
