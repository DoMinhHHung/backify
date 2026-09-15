package handler

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"

	"backify/services/control-plane/internal/domain"
	"backify/services/control-plane/internal/usecase"
)

type createEntityRequest struct {
	Name string `json:"name"`
}

type entityResponse struct {
	ID        string `json:"id"`
	ProjectID string `json:"project_id"`
	Name      string `json:"name"`
	IsSystem  bool   `json:"is_system"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

func toEntityResponse(e *domain.Entity) entityResponse {
	return entityResponse{
		ID:        e.ID,
		ProjectID: e.ProjectID,
		Name:      e.Name,
		IsSystem:  e.IsSystem,
		CreatedAt: e.CreatedAt.Format(time.RFC3339),
		UpdatedAt: e.UpdatedAt.Format(time.RFC3339),
	}
}

// AddEntityHandler tạo entity trong project từ nội dung JSON của yêu cầu.
func (h *Handler) AddEntityHandler(w http.ResponseWriter, r *http.Request) {
	projectID := chi.URLParam(r, "id")

	var req createEntityRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, domain.ErrInvalidInput.WithDetails(map[string]interface{}{"reason": "invalid json body"}))
		return
	}

	entity, err := h.AddEntity.Execute(r.Context(), usecase.AddEntityInput{
		ProjectID: projectID,
		Name:      req.Name,
	})
	if err != nil {
		writeError(w, err)
		return
	}

	writeData(w, http.StatusCreated, toEntityResponse(entity))
}

// ListEntitiesHandler trả về danh sách entity của project.
func (h *Handler) ListEntitiesHandler(w http.ResponseWriter, r *http.Request) {
	projectID := chi.URLParam(r, "id")

	entities, err := h.AddEntity.List(r.Context(), projectID)
	if err != nil {
		writeError(w, err)
		return
	}

	responses := make([]entityResponse, len(entities))
	for i, e := range entities {
		responses[i] = toEntityResponse(e)
	}

	writeData(w, http.StatusOK, responses)
}
