package domain

import (
	"net/mail"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"
)

const (
	MinPasswordLength = 8
	MaxPasswordLength = 72
	MaxFullNameLength = 100
	MaxPhoneLength    = 20
)

var phonePattern = regexp.MustCompile(`^[+]?[0-9]{7,15}$`)

type User struct {
	ID           string
	ProjectID    string
	Email        string
	PasswordHash string
	FullName     string
	Phone        string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

func validateEmail(email string) error {
	trimmed := strings.TrimSpace(strings.ToLower(email))
	if trimmed == "" {
		return ErrInvalidInput.WithDetails(map[string]interface{}{
			"field":  "email",
			"reason": "must not be empty",
		})
	}
	if _, err := mail.ParseAddress(trimmed); err != nil {
		return ErrInvalidInput.WithDetails(map[string]interface{}{
			"field":  "email",
			"reason": "invalid email format",
		})
	}
	return nil
}

func ValidatePassword(password string) error {
	length := utf8.RuneCountInString(password)
	if length < MinPasswordLength {
		return ErrInvalidInput.WithDetails(map[string]interface{}{
			"field":  "password",
			"reason": "must be at least 8 characters",
		})
	}
	if length > MaxPasswordLength {
		return ErrInvalidInput.WithDetails(map[string]interface{}{
			"field":  "password",
			"reason": "must be at most 72 characters",
		})
	}
	return nil
}

func validateFullName(name string) error {
	trimmed := strings.TrimSpace(name)
	if trimmed == "" {
		return nil
	}
	if utf8.RuneCountInString(trimmed) > MaxFullNameLength {
		return ErrInvalidInput.WithDetails(map[string]interface{}{
			"field":  "fullName",
			"reason": "must be at most 100 characters",
		})
	}
	return nil
}

func validatePhone(phone string) error {
	trimmed := strings.TrimSpace(phone)
	if trimmed == "" {
		return nil
	}
	if utf8.RuneCountInString(trimmed) > MaxPhoneLength {
		return ErrInvalidInput.WithDetails(map[string]interface{}{
			"field":  "phone",
			"reason": "must be at most 20 characters",
		})
	}
	if !phonePattern.MatchString(trimmed) {
		return ErrInvalidInput.WithDetails(map[string]interface{}{
			"field":  "phone",
			"reason": "must be 7-15 digits, optional leading +",
		})
	}
	return nil
}

func NewUser(projectID, email, fullName, phone string) (*User, error) {
	if strings.TrimSpace(projectID) == "" {
		return nil, ErrInvalidInput.WithDetails(map[string]interface{}{
			"field":  "project_id",
			"reason": "must not be empty",
		})
	}

	if err := validateEmail(email); err != nil {
		return nil, err
	}
	if err := validateFullName(fullName); err != nil {
		return nil, err
	}
	if err := validatePhone(phone); err != nil {
		return nil, err
	}

	id, err := generateID()
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	return &User{
		ID:        id,
		ProjectID: projectID,
		Email:     strings.TrimSpace(strings.ToLower(email)),
		FullName:  strings.TrimSpace(fullName),
		Phone:     strings.TrimSpace(phone),
		CreatedAt: now,
		UpdatedAt: now,
	}, nil
}

func (u *User) SetPasswordHash(hash string) error {
	if hash == "" {
		return ErrInvalidInput.WithDetails(map[string]interface{}{
			"field":  "password_hash",
			"reason": "must not be empty",
		})
	}
	u.PasswordHash = hash
	u.UpdatedAt = time.Now().UTC()
	return nil
}
