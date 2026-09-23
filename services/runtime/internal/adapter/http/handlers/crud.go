package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"github.com/DoMinhHHung/backify/services/runtime/internal/adapter/http/middleware"
	"github.com/DoMinhHHung/backify/services/runtime/internal/domain"
	"github.com/DoMinhHHung/backify/services/runtime/internal/usecase"
)

type CRUDHandler struct {
	crud *usecase.CRUD
}

func NewCRUDHandler(crud *usecase.CRUD) *CRUDHandler {
	return &CRUDHandler{crud: crud}
}

// Create tạo bản ghi cho entity trong URL bằng project config và claims của request.
func (h *CRUDHandler) Create(w http.ResponseWriter, r *http.Request) {
	cfg, claims, ok := ctxAuth(w, r)
	if !ok {
		return
	}
	entity := chi.URLParam(r, "entity")

	var body map[string]any
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}

	rec, err := h.crud.Create(r.Context(), usecase.CreateInput{
		ProjectConfig: cfg,
		Claims:        claims,
		EntityName:    entity,
		Data:          body,
	})
	if err != nil {
		writeDomainError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, rec)
}

// List liệt kê bản ghi của entity với limit và offset lấy từ query string.
func (h *CRUDHandler) List(w http.ResponseWriter, r *http.Request) {
	cfg, claims, ok := ctxAuth(w, r)
	if !ok {
		return
	}
	entity := chi.URLParam(r, "entity")
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	out, err := h.crud.List(r.Context(), usecase.ListInput{
		ProjectConfig: cfg,
		Claims:        claims,
		EntityName:    entity,
		Limit:         limit,
		Offset:        offset,
	})
	if err != nil {
		writeDomainError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, out)
}

// Get trả một bản ghi theo entity và id trong URL sau khi kiểm tra quyền đọc.
func (h *CRUDHandler) Get(w http.ResponseWriter, r *http.Request) {
	cfg, claims, ok := ctxAuth(w, r)
	if !ok {
		return
	}
	entity := chi.URLParam(r, "entity")
	id := chi.URLParam(r, "id")

	rec, err := h.crud.Get(r.Context(), usecase.GetInput{
		ProjectConfig: cfg,
		Claims:        claims,
		EntityName:    entity,
		ID:            id,
	})
	if err != nil {
		writeDomainError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, rec)
}

// Update cập nhật các trường được phép của bản ghi theo entity và id trong URL.
func (h *CRUDHandler) Update(w http.ResponseWriter, r *http.Request) {
	cfg, claims, ok := ctxAuth(w, r)
	if !ok {
		return
	}
	entity := chi.URLParam(r, "entity")
	id := chi.URLParam(r, "id")

	var body map[string]any
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}

	rec, err := h.crud.Update(r.Context(), usecase.UpdateInput{
		ProjectConfig: cfg,
		Claims:        claims,
		EntityName:    entity,
		ID:            id,
		Data:          body,
	})
	if err != nil {
		writeDomainError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, rec)
}

// Delete xóa bản ghi được phép truy cập và trả trạng thái 204 khi thành công.
func (h *CRUDHandler) Delete(w http.ResponseWriter, r *http.Request) {
	cfg, claims, ok := ctxAuth(w, r)
	if !ok {
		return
	}
	entity := chi.URLParam(r, "entity")
	id := chi.URLParam(r, "id")

	if err := h.crud.Delete(r.Context(), usecase.DeleteInput{
		ProjectConfig: cfg,
		Claims:        claims,
		EntityName:    entity,
		ID:            id,
	}); err != nil {
		writeDomainError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// ctxAuth lấy project config và claims từ context; khi thiếu dữ liệu, hàm tự ghi
// phản hồi lỗi tương ứng và trả false.
func ctxAuth(w http.ResponseWriter, r *http.Request) (*domain.ProjectConfig, *domain.AuthClaims, bool) {
	cfg, ok := domain.ProjectConfigFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusInternalServerError, "project context missing")
		return nil, nil, false
	}
	claims, ok := middleware.AuthClaimsFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return nil, nil, false
	}
	return cfg, claims, true
}
