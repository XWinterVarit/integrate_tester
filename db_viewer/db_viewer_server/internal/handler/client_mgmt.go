package handler

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/XWinterVarit/integrate_tester/db_viewer/db_viewer_server/internal/model"
	"github.com/XWinterVarit/integrate_tester/db_viewer/db_viewer_server/internal/service"
)

type ClientMgmtHandler struct {
	svc *service.ClientService
}

func NewClientMgmtHandler(svc *service.ClientService) *ClientMgmtHandler {
	return &ClientMgmtHandler{svc: svc}
}

func (h *ClientMgmtHandler) ListClients(w http.ResponseWriter, r *http.Request) {
	result, err := h.svc.ListClients(r.Context())
	if err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if result == nil {
		result = []model.ClientConfigResponse{}
	}
	writeJSON(w, result)
}

func (h *ClientMgmtHandler) CreateClient(w http.ResponseWriter, r *http.Request) {
	var req model.SaveClientRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, "invalid body", http.StatusBadRequest)
		return
	}
	if err := h.svc.SaveClient(r.Context(), req, true); err != nil {
		if strings.HasPrefix(err.Error(), "CONFLICT:") {
			writeError(w, strings.TrimPrefix(err.Error(), "CONFLICT:"), http.StatusConflict)
			return
		}
		writeError(w, err.Error(), http.StatusBadRequest)
		return
	}
	writeJSON(w, model.StatusResponse{Status: "ok"})
}

func (h *ClientMgmtHandler) UpdateClient(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")
	var req model.SaveClientRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, "invalid body", http.StatusBadRequest)
		return
	}
	req.Name = name
	if err := h.svc.SaveClient(r.Context(), req, false); err != nil {
		writeError(w, err.Error(), http.StatusBadRequest)
		return
	}
	writeJSON(w, model.StatusResponse{Status: "ok"})
}

func (h *ClientMgmtHandler) DeleteClient(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")
	if err := h.svc.DeleteClient(r.Context(), name); err != nil {
		writeError(w, err.Error(), http.StatusNotFound)
		return
	}
	writeJSON(w, model.StatusResponse{Status: "ok"})
}

func (h *ClientMgmtHandler) TestConnection(w http.ResponseWriter, r *http.Request) {
	var req model.TestConnectionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, "invalid body", http.StatusBadRequest)
		return
	}
	if err := h.svc.TestConnection(r.Context(), req); err != nil {
		writeJSON(w, model.TestConnectionResponse{Success: false, Error: err.Error()})
		return
	}
	writeJSON(w, model.TestConnectionResponse{Success: true})
}

func (h *ClientMgmtHandler) ListAllTables(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")
	tables, err := h.svc.ListAllTables(r.Context(), name)
	if err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, model.ListTablesFromDBResponse{Tables: tables})
}

func (h *ClientMgmtHandler) ReorderClients(w http.ResponseWriter, r *http.Request) {
	var req model.ReorderClientsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, "invalid body", http.StatusBadRequest)
		return
	}
	if err := h.svc.ReorderClients(r.Context(), req.Names); err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, model.StatusResponse{Status: "ok"})
}

func (h *ClientMgmtHandler) ListTablesFromConnection(w http.ResponseWriter, r *http.Request) {
	var req model.TestConnectionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, "invalid body", http.StatusBadRequest)
		return
	}
	tables, err := h.svc.ListAllTablesFromConnection(r.Context(), req)
	if err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, model.ListTablesFromDBResponse{Tables: tables})
}
