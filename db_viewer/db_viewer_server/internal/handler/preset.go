package handler

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/XWinterVarit/integrate_tester/db_viewer/db_viewer_server/internal/model"
	"github.com/XWinterVarit/integrate_tester/db_viewer/db_viewer_server/internal/service"
)

type PresetHandler struct {
	svc *service.PresetService
}

func NewPresetHandler(svc *service.PresetService) *PresetHandler {
	return &PresetHandler{svc: svc}
}

// --- Preset Filters ---

func (h *PresetHandler) ListFilters(w http.ResponseWriter, r *http.Request) {
	client := r.PathValue("client")
	table := r.PathValue("table")
	result, err := h.svc.ListPresetFilters(r.Context(), client, table)
	if err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if result == nil {
		result = []model.PresetFilterResponse{}
	}
	writeJSON(w, result)
}

func (h *PresetHandler) SaveFilter(w http.ResponseWriter, r *http.Request) {
	client := r.PathValue("client")
	table := r.PathValue("table")
	var req model.SavePresetFilterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, "invalid body", http.StatusBadRequest)
		return
	}
	if err := h.svc.SavePresetFilter(r.Context(), client, table, req); err != nil {
		writeError(w, err.Error(), http.StatusBadRequest)
		return
	}
	writeJSON(w, model.StatusResponse{Status: "ok"})
}

func (h *PresetHandler) DeleteFilter(w http.ResponseWriter, r *http.Request) {
	client := r.PathValue("client")
	table := r.PathValue("table")
	name := r.PathValue("name")
	if err := h.svc.DeletePresetFilter(r.Context(), client, table, name); err != nil {
		writeError(w, err.Error(), http.StatusNotFound)
		return
	}
	writeJSON(w, model.StatusResponse{Status: "ok"})
}

// --- Preset Queries ---

func (h *PresetHandler) ListQueries(w http.ResponseWriter, r *http.Request) {
	client := r.PathValue("client")
	table := r.PathValue("table")
	result, err := h.svc.ListPresetQueries(r.Context(), client, table)
	if err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if result == nil {
		result = []model.PresetQueryResponse{}
	}
	writeJSON(w, result)
}

func (h *PresetHandler) SaveQuery(w http.ResponseWriter, r *http.Request) {
	client := r.PathValue("client")
	table := r.PathValue("table")
	var req model.SavePresetQueryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, "invalid body", http.StatusBadRequest)
		return
	}
	if err := h.svc.SavePresetQuery(r.Context(), client, table, req); err != nil {
		writeError(w, err.Error(), http.StatusBadRequest)
		return
	}
	writeJSON(w, model.StatusResponse{Status: "ok"})
}

func (h *PresetHandler) DeleteQuery(w http.ResponseWriter, r *http.Request) {
	client := r.PathValue("client")
	table := r.PathValue("table")
	name := r.PathValue("name")
	if err := h.svc.DeletePresetQuery(r.Context(), client, table, name); err != nil {
		writeError(w, err.Error(), http.StatusNotFound)
		return
	}
	writeJSON(w, model.StatusResponse{Status: "ok"})
}

func (h *PresetHandler) ResolveQuery(w http.ResponseWriter, r *http.Request) {
	client := r.PathValue("client")
	table := r.PathValue("table")
	name := r.PathValue("name")
	q, err := h.svc.ResolvePresetQuery(r.Context(), client, table, name)
	if err != nil {
		writeError(w, fmt.Sprintf("resolve error: %v", err), http.StatusBadRequest)
		return
	}
	writeJSON(w, q)
}

func (h *PresetHandler) ValidateQuery(w http.ResponseWriter, r *http.Request) {
	client := r.PathValue("client")
	table := r.PathValue("table")
	var req model.ValidateQueryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, "invalid body", http.StatusBadRequest)
		return
	}
	result := h.svc.ValidateQuery(r.Context(), client, table, req)
	writeJSON(w, result)
}

// --- Field Descriptions ---

func (h *PresetHandler) GetFieldDescriptions(w http.ResponseWriter, r *http.Request) {
	client := r.PathValue("client")
	table := r.PathValue("table")
	descs, err := h.svc.GetFieldDescriptions(r.Context(), client, table)
	if err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, descs)
}

func (h *PresetHandler) SaveFieldDescriptions(w http.ResponseWriter, r *http.Request) {
	client := r.PathValue("client")
	table := r.PathValue("table")
	var descs map[string]string
	if err := json.NewDecoder(r.Body).Decode(&descs); err != nil {
		writeError(w, "invalid body", http.StatusBadRequest)
		return
	}
	if err := h.svc.SaveFieldDescriptions(r.Context(), client, table, descs); err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, model.StatusResponse{Status: "ok"})
}
