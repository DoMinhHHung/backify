package handler

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"

	"backify/services/control-plane/internal/domain"
	"backify/services/control-plane/internal/usecase"
)

type Handler struct {
	CreateProject *usecase.CreateProject
	GetProject    *usecase.GetProject
	AddEntity     *usecase.AddEntity
	AddField      *usecase.AddField
	DeleteField   *usecase.DeleteField
	ConfigModule  *usecase.ConfigModule
}

func New(
	createProject *usecase.CreateProject,
	getProject *usecase.GetProject,
	addEntity *usecase.AddEntity,
	addField *usecase.AddField,
	deleteField *usecase.DeleteField,
	configModule *usecase.ConfigModule,
) *Handler {
	return &Handler{
		CreateProject: createProject,
		GetProject:    getProject,
		AddEntity:     addEntity,
		AddField:      addField,
		DeleteField:   deleteField,
		ConfigModule:  configModule,
	}
}

func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Route("/api/v1", func(r chi.Router) {
		r.Post("/projects", h.CreateProjectHandler)
		r.Get("/projects", h.ListProjectsHandler)
		r.Get("/projects/{id}", h.GetProjectHandler)
		r.Delete("/projects/{id}", h.DeleteProjectHandler)

		r.Post("/projects/{id}/entities", h.AddEntityHandler)
		r.Get("/projects/{id}/entities", h.ListEntitiesHandler)

		r.Post("/entities/{eid}/fields", h.AddFieldHandler)
		r.Get("/entities/{eid}/fields", h.ListFieldsHandler)
		r.Delete("/fields/{fid}", h.DeleteFieldHandler)

		r.Post("/projects/{id}/modules/{name}/functions/{fn}/toggle-field", h.ToggleFieldHandler)
		r.Get("/projects/{id}/modules", h.ListModulesHandler)
	})
}

type envelope struct {
	Data  interface{} `json:"data,omitempty"`
	Error *errorBody  `json:"error,omitempty"`
}

type errorBody struct {
	Code    string      `json:"code"`
	Message string      `json:"message"`
	Details interface{} `json:"details,omitempty"`
}

func writeData(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(envelope{Data: data})
}

func writeError(w http.ResponseWriter, err error) {
	status, body := mapError(err)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(envelope{Error: body})
}

func mapError(err error) (int, *errorBody) {
	domainErr, ok := err.(*domain.Error)
	if !ok {
		return http.StatusInternalServerError, &errorBody{
			Code:    "INTERNAL_ERROR",
			Message: "an unexpected error occurred",
		}
	}

	body := &errorBody{
		Code:    string(domainErr.Code),
		Message: domainErr.Message,
		Details: domainErr.Details,
	}

	switch domainErr.Code {
	case domain.CodeInvalidInput, domain.CodeInvalidFieldType, domain.CodeInvalidModuleName:
		return http.StatusBadRequest, body
	case domain.CodeProjectNotFound, domain.CodeEntityNotFound, domain.CodeFieldNotFound,
		domain.CodeModuleNotFound, domain.CodeFunctionNotFound:
		return http.StatusNotFound, body
	case domain.CodeSubdomainTaken, domain.CodeEntityNameTaken, domain.CodeFieldNameTaken,
		domain.CodeSystemEntityCannotDelete, domain.CodeSystemFieldCannotDelete, domain.CodeFieldInUse:
		return http.StatusConflict, body
	default:
		return http.StatusInternalServerError, body
	}
}

func decodeJSON(r *http.Request, dst interface{}) error {
	defer r.Body.Close()
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	return decoder.Decode(dst)
}
