package handler

import (
	"encoding/json"
	"net/http"

	"github.com/rs/zerolog/log"

	"backify/services/auth/internal/domain"
)

// envelope khớp convention control-plane — không tự nghĩ format mới.
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
	if err := json.NewEncoder(w).Encode(envelope{Data: data}); err != nil {
		log.Error().Err(err).Msg("failed to encode response")
	}
}

func writeError(w http.ResponseWriter, err error) {
	status, body := mapError(err)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if encErr := json.NewEncoder(w).Encode(envelope{Error: body}); encErr != nil {
		log.Error().Err(encErr).Msg("failed to encode error response")
	}
}

// mapError chuyển lỗi domain Auth thành HTTP status + body công khai.
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
	case domain.CodeInvalidInput:
		return http.StatusBadRequest, body
	case domain.CodeUserNotFound, domain.CodeProjectNotFound:
		return http.StatusNotFound, body
	case domain.CodeEmailTaken:
		return http.StatusConflict, body
	case domain.CodeInvalidCredentials:
		return http.StatusUnauthorized, body
	case domain.CodeTokenInvalid, domain.CodeTokenExpired, domain.CodeTokenRevoked, domain.CodeTokenReuseDetected:
		return http.StatusUnauthorized, body
	case domain.CodeProjectDatabaseMissing:
		// project.created event chưa xử lý xong — DB per-project chưa sẵn.
		return http.StatusServiceUnavailable, body
	default:
		return http.StatusInternalServerError, body
	}
}

// decodeJSON giải mã body và từ chối trường JSON không có trong kiểu đích.
func decodeJSON(r *http.Request, dst interface{}) error {
	defer r.Body.Close()
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	return decoder.Decode(dst)
}
