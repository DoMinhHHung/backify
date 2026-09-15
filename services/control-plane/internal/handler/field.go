package handler

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"

	"backify/services/control-plane/internal/domain"
	"backify/services/control-plane/internal/usecase"
)

type createFieldRequest struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

type fieldResponse struct {
	ID        string `json:"id"`
	EntityID  string `json:"entity_id"`
	Name      string `json:"name"`
	Type      string `json:"type"`
	IsSystem  bool   `json:"is_system"`
	CreatedAt string `json:"created_at"`
}

func toFieldResponse(f *domain.Field) fieldResponse {
	return fieldResponse{
		ID:        f.ID,
		EntityID:  f.EntityID,
		Name:      f.Name,
		Type:      string(f.Type),
		IsSystem:  f.IsSystem,
		CreatedAt: f.CreatedAt.Format(time.RFC3339),
	}
}

func (h *Handler) AddFieldHandler(w http.ResponseWriter, r *http.Request) {
	entityID := chi.URLParam(r, "eid")

	var req createFieldRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, domain.ErrInvalidInput.WithDetails(map[string]interface{}{"reason": "invalid json body"}))
		return
	}

	field, err := h.AddField.Execute(r.Context(), usecase.AddFieldInput{
		EntityID: entityID,
		Name:     req.Name,
		Type:     domain.FieldType(req.Type),
	})
	if err != nil {
		writeError(w, err)
		return
	}

	writeData(w, http.StatusCreated, toFieldResponse(field))
}

func (h *Handler) ListFieldsHandler(w http.ResponseWriter, r *http.Request) {
	entityID := chi.URLParam(r, "eid")

	fields, err := h.AddField.List(r.Context(), entityID)
	if err != nil {
		writeError(w, err)
		return
	}

	responses := make([]fieldResponse, len(fields))
	for i, f := range fields {
		responses[i] = toFieldResponse(f)
	}

	writeData(w, http.StatusOK, responses)
}

func (h *Handler) DeleteFieldHandler(w http.ResponseWriter, r *http.Request) {
	fieldID := chi.URLParam(r, "fid")
	force := r.URL.Query().Get("force") == "true"

	err := h.DeleteField.Execute(r.Context(), usecase.DeleteFieldInput{
		FieldID: fieldID,
		Force:   force,
	})
	if err != nil {
		writeError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
