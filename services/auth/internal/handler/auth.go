package handler

import (
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"

	"backify/services/auth/internal/domain"
	"backify/services/auth/internal/usecase"
)

// Handler gắn HTTP routes cho Auth Service.
type Handler struct {
	SignUp         *usecase.SignUp
	SignIn         *usecase.SignIn
	SignOut        *usecase.SignOut
	RefreshToken   *usecase.RefreshToken
	ForgotPassword *usecase.ForgotPassword
	ResetPassword  *usecase.ResetPassword
	Me             *usecase.Me
	VerifyToken    *usecase.VerifyToken
}

// New tạo Handler từ các usecase đã wire.
func New(
	signUp *usecase.SignUp,
	signIn *usecase.SignIn,
	signOut *usecase.SignOut,
	refresh *usecase.RefreshToken,
	forgot *usecase.ForgotPassword,
	reset *usecase.ResetPassword,
	me *usecase.Me,
	verify *usecase.VerifyToken,
) *Handler {
	return &Handler{
		SignUp:         signUp,
		SignIn:         signIn,
		SignOut:        signOut,
		RefreshToken:   refresh,
		ForgotPassword: forgot,
		ResetPassword:  reset,
		Me:             me,
		VerifyToken:    verify,
	}
}

// RegisterRoutes mount /auth/* — không có prefix /api/v1 (đúng spec gốc).
func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Route("/auth", func(r chi.Router) {
		r.Use(RequireProjectID)
		r.Post("/signup", h.SignUpHandler)
		r.Post("/signin", h.SignInHandler)
		r.Post("/refresh", h.RefreshHandler)
		r.Post("/forgot-password", h.ForgotPasswordHandler)
		r.Post("/reset-password", h.ResetPasswordHandler)

		r.Group(func(r chi.Router) {
			r.Use(RequireAuth(h.VerifyToken))
			r.Post("/signout", h.SignOutHandler)
			r.Get("/me", h.MeHandler)
		})
	})
}

// --- request / response DTOs ---

type signUpRequest struct {
	Email           string `json:"email"`
	Password        string `json:"password"`
	FullName        string `json:"fullName"`
	Phone           string `json:"phone"`
	RefreshDuration string `json:"refreshDuration"`
}

type signInRequest struct {
	Email           string `json:"email"`
	Password        string `json:"password"`
	RefreshDuration string `json:"refreshDuration"`
}

type refreshRequest struct {
	RefreshToken string `json:"refreshToken"`
}

type forgotPasswordRequest struct {
	Email string `json:"email"`
}

type resetPasswordRequest struct {
	ResetToken  string `json:"resetToken"`
	NewPassword string `json:"newPassword"`
}

type signOutRequest struct {
	RefreshToken string `json:"refreshToken"`
}

type authTokensResponse struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
	ExpiresIn    int64  `json:"expiresIn"`
}

type userResponse struct {
	ID       string `json:"id"`
	Email    string `json:"email"`
	FullName string `json:"fullName"`
	Phone    string `json:"phone"`
}

type signUpResponse struct {
	User   userResponse       `json:"user"`
	Tokens authTokensResponse `json:"tokens"`
}

type signInResponse struct {
	User   userResponse       `json:"user"`
	Tokens authTokensResponse `json:"tokens"`
}

func toUserResponse(u *domain.User) userResponse {
	return userResponse{
		ID:       u.ID,
		Email:    u.Email,
		FullName: u.FullName,
		Phone:    u.Phone,
	}
}

func toTokensResponse(t usecase.AuthTokens) authTokensResponse {
	return authTokensResponse{
		AccessToken:  t.AccessToken,
		RefreshToken: t.RefreshToken,
		ExpiresIn:    t.ExpiresIn,
	}
}

// --- handlers ---

func (h *Handler) SignUpHandler(w http.ResponseWriter, r *http.Request) {
	var req signUpRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, domain.ErrInvalidInput.WithDetails(map[string]interface{}{
			"reason": "invalid JSON body",
		}))
		return
	}
	out, err := h.SignUp.Execute(r.Context(), usecase.SignUpInput{
		ProjectID:       projectIDFrom(r.Context()),
		Email:           req.Email,
		Password:        req.Password,
		FullName:        req.FullName,
		Phone:           req.Phone,
		RefreshDuration: req.RefreshDuration,
	})
	if err != nil {
		writeError(w, err)
		return
	}
	writeData(w, http.StatusCreated, signUpResponse{
		User:   toUserResponse(out.User),
		Tokens: toTokensResponse(out.Tokens),
	})
}

func (h *Handler) SignInHandler(w http.ResponseWriter, r *http.Request) {
	var req signInRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, domain.ErrInvalidInput.WithDetails(map[string]interface{}{
			"reason": "invalid JSON body",
		}))
		return
	}
	out, err := h.SignIn.Execute(r.Context(), usecase.SignInInput{
		ProjectID:       projectIDFrom(r.Context()),
		Email:           req.Email,
		Password:        req.Password,
		RefreshDuration: req.RefreshDuration,
	})
	if err != nil {
		writeError(w, err)
		return
	}
	writeData(w, http.StatusOK, signInResponse{
		User:   toUserResponse(out.User),
		Tokens: toTokensResponse(out.Tokens),
	})
}

func (h *Handler) RefreshHandler(w http.ResponseWriter, r *http.Request) {
	var req refreshRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, domain.ErrInvalidInput.WithDetails(map[string]interface{}{
			"reason": "invalid JSON body",
		}))
		return
	}
	out, err := h.RefreshToken.Execute(r.Context(), usecase.RefreshTokenInput{
		ProjectID:    projectIDFrom(r.Context()),
		RefreshToken: req.RefreshToken,
	})
	if err != nil {
		writeError(w, err)
		return
	}
	writeData(w, http.StatusOK, toTokensResponse(out.Tokens))
}

func (h *Handler) ForgotPasswordHandler(w http.ResponseWriter, r *http.Request) {
	var req forgotPasswordRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, domain.ErrInvalidInput.WithDetails(map[string]interface{}{
			"reason": "invalid JSON body",
		}))
		return
	}
	if err := h.ForgotPassword.Execute(r.Context(), usecase.ForgotPasswordInput{
		ProjectID: projectIDFrom(r.Context()),
		Email:     req.Email,
	}); err != nil {
		writeError(w, err)
		return
	}
	// Luôn 200 — kể cả email không tồn tại (anti-enumeration).
	writeData(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *Handler) ResetPasswordHandler(w http.ResponseWriter, r *http.Request) {
	var req resetPasswordRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, domain.ErrInvalidInput.WithDetails(map[string]interface{}{
			"reason": "invalid JSON body",
		}))
		return
	}
	if err := h.ResetPassword.Execute(r.Context(), usecase.ResetPasswordInput{
		ProjectID:   projectIDFrom(r.Context()),
		ResetToken:  req.ResetToken,
		NewPassword: req.NewPassword,
	}); err != nil {
		writeError(w, err)
		return
	}
	writeData(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *Handler) SignOutHandler(w http.ResponseWriter, r *http.Request) {
	var req signOutRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, domain.ErrInvalidInput.WithDetails(map[string]interface{}{
			"reason": "invalid JSON body",
		}))
		return
	}
	// Access token raw từ header — usecase cần token string, không phải claims.
	auth := r.Header.Get("Authorization")
	accessToken := ""
	if strings.HasPrefix(auth, "Bearer ") {
		accessToken = strings.TrimSpace(auth[len("Bearer "):])
	}
	if err := h.SignOut.Execute(r.Context(), usecase.SignOutInput{
		ProjectID:    projectIDFrom(r.Context()),
		AccessToken:  accessToken,
		RefreshToken: req.RefreshToken,
	}); err != nil {
		writeError(w, err)
		return
	}
	writeData(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *Handler) MeHandler(w http.ResponseWriter, r *http.Request) {
	user, err := h.Me.Execute(r.Context(), projectIDFrom(r.Context()), userIDFrom(r.Context()))
	if err != nil {
		writeError(w, err)
		return
	}
	writeData(w, http.StatusOK, toUserResponse(user))
}
