package port

import (
	"github.com/DoMinhHHung/backify/services/runtime/internal/domain"
)

type TokenService interface {
	Issue(userID, projectID, role string) (*domain.TokenPair, error)
	ParseAccess(token string) (*domain.AuthClaims, error)
	ParseRefresh(token string) (*domain.AuthClaims, error)
}
