package handlers

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/DoMinhHHung/backify/services/runtime/internal/adapter/http/middleware"
	"github.com/DoMinhHHung/backify/services/runtime/internal/domain"
	"github.com/DoMinhHHung/backify/services/runtime/internal/usecase"
)

type AuthHandler struct {
	auth *usecase.Auth
}

func NewAuthHandler(auth *usecase.Auth) *AuthHandler {
	return &AuthHandler{auth: auth}
}

// Signup đăng ký người dùng trong dự án của request và trả người dùng cùng cặp token.
func (h *AuthHandler) Signup(w http.ResponseWriter, r *http.Request) {
	cfg, ok := domain.ProjectConfigFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusInternalServerError, "project context missing")
		return
	}

	var body map[string]string
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}

	out, err := h.auth.Signup(r.Context(), usecase.SignupInput{
		ProjectConfig: cfg,
		Fields:        body,
	})
	if err != nil {
		WriteDomainError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{
		"user":   publicUser(out.User),
		"tokens": out.Tokens,
	})
}

// Signin xác thực email và mật khẩu trong dự án của request rồi trả người dùng cùng cặp token.
func (h *AuthHandler) Signin(w http.ResponseWriter, r *http.Request) {
	cfg, ok := domain.ProjectConfigFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusInternalServerError, "project context missing")
		return
	}

	var body struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}

	out, err := h.auth.Signin(r.Context(), usecase.SigninInput{
		ProjectConfig: cfg,
		Email:         body.Email,
		Password:      body.Password,
	})
	if err != nil {
		WriteDomainError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"user":   publicUser(out.User),
		"tokens": out.Tokens,
	})
}

// Refresh đổi refresh token hợp lệ của dự án hiện tại thành một cặp token mới.
func (h *AuthHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	cfg, ok := domain.ProjectConfigFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusInternalServerError, "project context missing")
		return
	}

	var body struct {
		RefreshToken string `json:"refreshToken"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}

	tokens, err := h.auth.Refresh(r.Context(), usecase.RefreshInput{
		ProjectConfig: cfg,
		RefreshToken:  body.RefreshToken,
	})
	if err != nil {
		WriteDomainError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, tokens)
}

// Me trả thông tin người dùng từ claims đã được middleware xác thực.
func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	cfg, ok := domain.ProjectConfigFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusInternalServerError, "project context missing")
		return
	}
	claims, ok := middleware.AuthClaimsFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	user, err := h.auth.Me(r.Context(), usecase.MeInput{
		ProjectConfig: cfg,
		UserID:        claims.UserID,
	})
	if err != nil {
		WriteDomainError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, publicUser(user))
}

// publicUser chuyển User thành payload công khai và loại bỏ password hash.
func publicUser(u *domain.User) map[string]any {
	return map[string]any{
		"id":        u.ID,
		"email":     u.Email,
		"fullName":  u.FullName,
		"phone":     u.Phone,
		"role":      u.Role,
		"createdAt": u.CreatedAt,
		"updatedAt": u.UpdatedAt,
	}
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

// WriteDomainError ánh xạ DomainError sang mã HTTP và payload JSON; lỗi khác được
// ghi log rồi chuyển thành phản hồi 500 không chứa chi tiết nội bộ.
func WriteDomainError(w http.ResponseWriter, err error) {
	if de, ok := err.(*domain.DomainError); ok {
		status := http.StatusBadRequest
		switch de.Code {
		case "unauthorized":
			status = http.StatusUnauthorized
		case "forbidden":
			status = http.StatusForbidden
		case "not_found":
			status = http.StatusNotFound
		case "email_taken":
			status = http.StatusConflict
		case "invalid_credentials":
			status = http.StatusUnauthorized
		}
		writeJSON(w, status, map[string]string{"error": de.Message, "code": de.Code})
		return
	}
	log.Printf("auth internal error: %v", err)
	log.Printf("handler internal error: %v", err)
	writeError(w, http.StatusInternalServerError, "internal error")
}

func writeDomainError(w http.ResponseWriter, err error) {
	WriteDomainError(w, err)
}
