package handler

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"

	"backify/services/control-plane/internal/domain"
	"backify/services/control-plane/internal/usecase"
)

type toggleFieldRequest struct {
	FieldID string `json:"field_id"`
	Enabled bool   `json:"enabled"`
}

type moduleResponse struct {
	ID        string `json:"id"`
	ProjectID string `json:"project_id"`
	Name      string `json:"name"`
	CreatedAt string `json:"created_at"`
}

func toModuleResponse(m *domain.Module) moduleResponse {
	return moduleResponse{
		ID:        m.ID,
		ProjectID: m.ProjectID,
		Name:      string(m.Name),
		CreatedAt: m.CreatedAt.Format(time.RFC3339),
	}
}

// ToggleFieldHandler bật hoặc tắt một field cho hàm của module.
func (h *Handler) ToggleFieldHandler(w http.ResponseWriter, r *http.Request) {
	projectID := chi.URLParam(r, "id")
	moduleName := chi.URLParam(r, "name")
	function := chi.URLParam(r, "fn")

	var req toggleFieldRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, domain.ErrInvalidInput.WithDetails(map[string]interface{}{"reason": "invalid json body"}))
		return
	}

	err := h.ConfigModule.ToggleField(r.Context(), usecase.ToggleFieldInput{
		ProjectID: projectID,
		Module:    domain.ModuleName(moduleName),
		Function:  function,
		FieldID:   req.FieldID,
		Enabled:   req.Enabled,
	})
	if err != nil {
		writeError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// ListModulesHandler trả về các module đã được cấu hình cho project.
func (h *Handler) ListModulesHandler(w http.ResponseWriter, r *http.Request) {
	projectID := chi.URLParam(r, "id")

	modules, err := h.ConfigModule.ListModules(r.Context(), projectID)
	if err != nil {
		writeError(w, err)
		return
	}

	responses := make([]moduleResponse, len(modules))
	for i, m := range modules {
		responses[i] = toModuleResponse(m)
	}

	writeData(w, http.StatusOK, responses)
}
