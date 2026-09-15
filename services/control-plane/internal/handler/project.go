package handler

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"

	"backify/services/control-plane/internal/domain"
	"backify/services/control-plane/internal/usecase"
)

type createProjectRequest struct {
	Name      string `json:"name"`
	Subdomain string `json:"subdomain"`
}

type projectResponse struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	Subdomain  string `json:"subdomain"`
	SchemaName string `json:"schema_name"`
	Plan       string `json:"plan"`
	Status     string `json:"status"`
	CreatedAt  string `json:"created_at"`
	UpdatedAt  string `json:"updated_at"`
}

func toProjectResponse(p *domain.Project) projectResponse {
	return projectResponse{
		ID:         p.ID,
		Name:       p.Name,
		Subdomain:  p.Subdomain,
		SchemaName: p.SchemaName,
		Plan:       string(p.Plan),
		Status:     string(p.Status),
		CreatedAt:  p.CreatedAt.Format(time.RFC3339),
		UpdatedAt:  p.UpdatedAt.Format(time.RFC3339),
	}
}

func (h *Handler) CreateProjectHandler(w http.ResponseWriter, r *http.Request) {
	var req createProjectRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, domain.ErrInvalidInput.WithDetails(map[string]interface{}{"reason": "invalid json body"}))
		return
	}

	project, err := h.CreateProject.Execute(r.Context(), usecase.CreateProjectInput{
		Name:      req.Name,
		Subdomain: req.Subdomain,
	})
	if err != nil {
		writeError(w, err)
		return
	}

	writeData(w, http.StatusCreated, toProjectResponse(project))
}

func (h *Handler) ListProjectsHandler(w http.ResponseWriter, r *http.Request) {
	projects, err := h.GetProject.List(r.Context())
	if err != nil {
		writeError(w, err)
		return
	}

	responses := make([]projectResponse, len(projects))
	for i, p := range projects {
		responses[i] = toProjectResponse(p)
	}

	writeData(w, http.StatusOK, responses)
}

func (h *Handler) GetProjectHandler(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	project, err := h.GetProject.Execute(r.Context(), id)
	if err != nil {
		writeError(w, err)
		return
	}

	writeData(w, http.StatusOK, toProjectResponse(project))
}

func (h *Handler) DeleteProjectHandler(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	if err := h.GetProject.Delete(r.Context(), id); err != nil {
		writeError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
