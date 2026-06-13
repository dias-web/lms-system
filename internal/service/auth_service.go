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
	CreateUser(ctx context.Context, in keycloak.CreateUserInput) (string, error)
	GetUser(ctx context.Context, userID string) (*keycloak.User, error)
	UpdateUser(ctx context.Context, userID string, in keycloak.UpdateUserInput) error
	SetPassword(ctx context.Context, userID, newPassword string) error
}

type AuthService interface {
	Login(ctx context.Context, req dto.LoginRequest) (dto.TokenResponse, error)
	Refresh(ctx context.Context, req dto.RefreshRequest) (dto.TokenResponse, error)
	Register(ctx context.Context, req dto.RegisterRequest) (dto.UserResponse, error)
	UpdateProfile(ctx context.Context, userID string, req dto.UpdateProfileRequest) (dto.UserResponse, error)
	ChangePassword(ctx context.Context, userID, username string, req dto.ChangePasswordRequest) error
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

func (s *authService) Register(ctx context.Context, req dto.RegisterRequest) (dto.UserResponse, error) {
	s.log.WithField("username", req.Username).Info("Registering new user")
	s.log.WithFields(logrus.Fields{
		"username": req.Username,
		"email":    req.Email,
		"role":     req.Role,
	}).Debug("user registration payload")

	id, err := s.kc.CreateUser(ctx, keycloak.CreateUserInput{
		Username:  req.Username,
		Email:     req.Email,
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Password:  req.Password,
		Role:      req.Role,
	})
	if err != nil {
		if errors.Is(err, keycloak.ErrUserExists) {
			s.log.WithField("username", req.Username).Warn("registration failed: user already exists")
			return dto.UserResponse{}, fmt.Errorf("%w: username or email already taken", ErrConflict)
		}
		s.log.WithError(err).WithField("username", req.Username).Error("registration failed: keycloak error")
		return dto.UserResponse{}, err
	}

	s.log.WithFields(logrus.Fields{"user_id": id, "username": req.Username, "role": req.Role}).
		Info("user registered")
	return dto.UserResponse{
		ID:        id,
		Username:  req.Username,
		Email:     req.Email,
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Role:      req.Role,
	}, nil
}

func (s *authService) UpdateProfile(ctx context.Context, userID string, req dto.UpdateProfileRequest) (dto.UserResponse, error) {
	s.log.WithField("user_id", userID).Info("Updating user profile")
	s.log.WithFields(logrus.Fields{
		"user_id":    userID,
		"email_set":  req.Email != "",
		"name_fields": req.FirstName != "" || req.LastName != "",
	}).Debug("profile update payload")

	if err := s.kc.UpdateUser(ctx, userID, keycloak.UpdateUserInput{
		Email:     req.Email,
		FirstName: req.FirstName,
		LastName:  req.LastName,
	}); err != nil {
		return dto.UserResponse{}, mapUserManagementErr(s, userID, err)
	}

	// Return the fresh server-side state, including the unchanged role.
	user, err := s.kc.GetUser(ctx, userID)
	if err != nil {
		s.log.WithError(err).WithField("user_id", userID).Error("profile updated but reload failed")
		return dto.UserResponse{}, err
	}

	s.log.WithField("user_id", userID).Info("user profile updated")
	return dto.UserResponse{
		ID:        user.ID,
		Username:  user.Username,
		Email:     user.Email,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Role:      user.Role,
	}, nil
}

func (s *authService) ChangePassword(ctx context.Context, userID, username string, req dto.ChangePasswordRequest) error {
	s.log.WithField("user_id", userID).Info("Changing user password")

	// Verify the current password by attempting a login, so a stolen access
	// token alone cannot reset the password.
	if _, err := s.kc.Login(ctx, username, req.OldPassword); err != nil {
		if errors.Is(err, keycloak.ErrInvalidCredentials) {
			s.log.WithField("user_id", userID).Warn("password change failed: wrong current password")
			return fmt.Errorf("%w: current password is incorrect", ErrUnauthorized)
		}
		s.log.WithError(err).WithField("user_id", userID).Error("password change failed: keycloak error")
		return err
	}

	if err := s.kc.SetPassword(ctx, userID, req.NewPassword); err != nil {
		s.log.WithError(err).WithField("user_id", userID).Error("password change failed: set password")
		return err
	}

	s.log.WithField("user_id", userID).Info("user password changed")
	return nil
}

// mapUserManagementErr translates not-found / conflict errors from the admin
// user API into domain errors, logging at the appropriate level.
func mapUserManagementErr(s *authService, userID string, err error) error {
	switch {
	case errors.Is(err, keycloak.ErrUserNotFound):
		s.log.WithField("user_id", userID).Warn("user not found")
		return fmt.Errorf("%w: user not found", ErrNotFound)
	case errors.Is(err, keycloak.ErrUserExists):
		s.log.WithField("user_id", userID).Warn("email already taken")
		return fmt.Errorf("%w: email already taken", ErrConflict)
	default:
		s.log.WithError(err).WithField("user_id", userID).Error("user management failed")
		return err
	}
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