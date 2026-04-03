package handler

import (
	"net/http"

	"github.com/XWinterVarit/integrate_tester/db_viewer/db_viewer_server/internal/service"
)

type ClientHandler struct {
	svc *service.TableService
}

func NewClientHandler(svc *service.TableService) *ClientHandler {
	return &ClientHandler{svc: svc}
}

func (h *ClientHandler) List(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, h.svc.ListClients())
}

func (h *ClientHandler) ListTables(w http.ResponseWriter, r *http.Request) {
	client := r.PathValue("client")
	tables, err := h.svc.ListTables(client)
	if err != nil {
		writeError(w, err.Error(), http.StatusNotFound)
		return
	}
	writeJSON(w, tables)
}
