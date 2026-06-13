package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/dias-web/lms-system/internal/dto"
	"github.com/dias-web/lms-system/internal/keycloak"
	"github.com/sirupsen/logrus"
)

// KeycloakClient is the slice of Keycloak behaviour the auth service relies on.
// It is satisfied by *keycloak.Client and mocked in tests.
type KeycloakClient interface {
	Login(ctx context.Context, username, password string) (*keycloak.Token, error)
	Refresh(ctx context.Context, refreshToken string) (*keycloak.Token, error)
}

type AuthService interface {
	Login(ctx context.Context, req dto.LoginRequest) (dto.TokenResponse, error)
	Refresh(ctx context.Context, req dto.RefreshRequest) (dto.TokenResponse, error)
}

type authService struct {
	kc  KeycloakClient
	log *logrus.Logger
}

func NewAuthService(kc KeycloakClient, log *logrus.Logger) AuthService {
	return &authService{kc: kc, log: log}
}

func (s *authService) Login(ctx context.Context, req dto.LoginRequest) (dto.TokenResponse, error) {
	s.log.WithField("username", req.Username).Info("User login attempt")

	token, err := s.kc.Login(ctx, req.Username, req.Password)
	if err != nil {
		if errors.Is(err, keycloak.ErrInvalidCredentials) {
			s.log.WithField("username", req.Username).Warn("login failed: invalid credentials")
			return dto.TokenResponse{}, fmt.Errorf("%w: invalid username or password", ErrUnauthorized)
		}
		s.log.WithError(err).WithField("username", req.Username).Error("login failed: keycloak error")
		return dto.TokenResponse{}, err
	}

	s.log.WithField("username", req.Username).Info("User logged in")
	return toTokenResponse(token), nil
}

func (s *authService) Refresh(ctx context.Context, req dto.RefreshRequest) (dto.TokenResponse, error) {
	s.log.Info("Refreshing access token")

	token, err := s.kc.Refresh(ctx, req.RefreshToken)
	if err != nil {
		if errors.Is(err, keycloak.ErrInvalidCredentials) {
			s.log.Warn("refresh failed: invalid or expired refresh token")
			return dto.TokenResponse{}, fmt.Errorf("%w: invalid or expired refresh token", ErrUnauthorized)
		}
		s.log.WithError(err).Error("refresh failed: keycloak error")
		return dto.TokenResponse{}, err
	}

	s.log.Info("Access token refreshed")
	return toTokenResponse(token), nil
}

func toTokenResponse(t *keycloak.Token) dto.TokenResponse {
	return dto.TokenResponse{
		AccessToken:      t.AccessToken,
		RefreshToken:     t.RefreshToken,
		ExpiresIn:        t.ExpiresIn,
		RefreshExpiresIn: t.RefreshExpiresIn,
		TokenType:        t.TokenType,
	}
}