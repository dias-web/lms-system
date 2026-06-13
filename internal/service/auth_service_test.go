package service

import (
	"context"
	"errors"
	"testing"

	"github.com/dias-web/lms-system/internal/dto"
	"github.com/dias-web/lms-system/internal/keycloak"
	svcmocks "github.com/dias-web/lms-system/internal/service/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestAuthService_Login_Success(t *testing.T) {
	kc := svcmocks.NewMockKeycloakClient(t)
	svc := NewAuthService(kc, silentLogger())

	kc.EXPECT().Login(mock.Anything, "admin", "admin123").Return(&keycloak.Token{
		AccessToken:      "access.jwt",
		RefreshToken:     "refresh.jwt",
		ExpiresIn:        300,
		RefreshExpiresIn: 604800,
		TokenType:        "Bearer",
	}, nil)

	resp, err := svc.Login(context.Background(), dto.LoginRequest{
		Username: "admin", Password: "admin123",
	})
	require.NoError(t, err)
	assert.Equal(t, "access.jwt", resp.AccessToken)
	assert.Equal(t, "refresh.jwt", resp.RefreshToken)
	assert.Equal(t, 300, resp.ExpiresIn)
	assert.Equal(t, 604800, resp.RefreshExpiresIn)
	assert.Equal(t, "Bearer", resp.TokenType)
}

func TestAuthService_Login_InvalidCredentials(t *testing.T) {
	kc := svcmocks.NewMockKeycloakClient(t)
	svc := NewAuthService(kc, silentLogger())

	kc.EXPECT().Login(mock.Anything, "admin", "wrong").
		Return(nil, keycloak.ErrInvalidCredentials)

	_, err := svc.Login(context.Background(), dto.LoginRequest{
		Username: "admin", Password: "wrong",
	})
	assert.ErrorIs(t, err, ErrUnauthorized)
}

func TestAuthService_Login_KeycloakDown(t *testing.T) {
	kc := svcmocks.NewMockKeycloakClient(t)
	svc := NewAuthService(kc, silentLogger())

	boom := errors.New("connection refused")
	kc.EXPECT().Login(mock.Anything, "admin", "admin123").Return(nil, boom)

	_, err := svc.Login(context.Background(), dto.LoginRequest{
		Username: "admin", Password: "admin123",
	})
	assert.ErrorIs(t, err, boom)
	assert.NotErrorIs(t, err, ErrUnauthorized)
}

func TestAuthService_Refresh_Success(t *testing.T) {
	kc := svcmocks.NewMockKeycloakClient(t)
	svc := NewAuthService(kc, silentLogger())

	kc.EXPECT().Refresh(mock.Anything, "old.refresh").Return(&keycloak.Token{
		AccessToken:      "new.access",
		RefreshToken:     "new.refresh",
		ExpiresIn:        300,
		RefreshExpiresIn: 604800,
		TokenType:        "Bearer",
	}, nil)

	resp, err := svc.Refresh(context.Background(), dto.RefreshRequest{RefreshToken: "old.refresh"})
	require.NoError(t, err)
	assert.Equal(t, "new.access", resp.AccessToken)
	assert.Equal(t, "new.refresh", resp.RefreshToken)
	assert.Equal(t, 300, resp.ExpiresIn)
}

func TestAuthService_Refresh_Expired(t *testing.T) {
	kc := svcmocks.NewMockKeycloakClient(t)
	svc := NewAuthService(kc, silentLogger())

	kc.EXPECT().Refresh(mock.Anything, "expired.refresh").
		Return(nil, keycloak.ErrInvalidCredentials)

	_, err := svc.Refresh(context.Background(), dto.RefreshRequest{RefreshToken: "expired.refresh"})
	assert.ErrorIs(t, err, ErrUnauthorized)
}

func TestAuthService_Refresh_KeycloakDown(t *testing.T) {
	kc := svcmocks.NewMockKeycloakClient(t)
	svc := NewAuthService(kc, silentLogger())

	boom := errors.New("connection refused")
	kc.EXPECT().Refresh(mock.Anything, "x").Return(nil, boom)

	_, err := svc.Refresh(context.Background(), dto.RefreshRequest{RefreshToken: "x"})
	assert.ErrorIs(t, err, boom)
	assert.NotErrorIs(t, err, ErrUnauthorized)
}

func TestAuthService_Register_Success(t *testing.T) {
	kc := svcmocks.NewMockKeycloakClient(t)
	svc := NewAuthService(kc, silentLogger())

	req := dto.RegisterRequest{
		Username: "jdoe", Email: "jdoe@lms.local", Password: "secret123",
		FirstName: "John", LastName: "Doe", Role: "ROLE_TEACHER",
	}
	kc.EXPECT().
		CreateUser(mock.Anything, keycloak.CreateUserInput{
			Username: "jdoe", Email: "jdoe@lms.local", FirstName: "John",
			LastName: "Doe", Password: "secret123", Role: "ROLE_TEACHER",
		}).
		Return("user-uuid-123", nil)

	resp, err := svc.Register(context.Background(), req)
	require.NoError(t, err)
	assert.Equal(t, "user-uuid-123", resp.ID)
	assert.Equal(t, "jdoe", resp.Username)
	assert.Equal(t, "ROLE_TEACHER", resp.Role)
}

func TestAuthService_Register_Conflict(t *testing.T) {
	kc := svcmocks.NewMockKeycloakClient(t)
	svc := NewAuthService(kc, silentLogger())

	kc.EXPECT().CreateUser(mock.Anything, mock.Anything).
		Return("", keycloak.ErrUserExists)

	_, err := svc.Register(context.Background(), dto.RegisterRequest{
		Username: "admin", Email: "admin@lms.local", Password: "x", Role: "ROLE_USER",
	})
	assert.ErrorIs(t, err, ErrConflict)
}

func TestAuthService_Register_KeycloakDown(t *testing.T) {
	kc := svcmocks.NewMockKeycloakClient(t)
	svc := NewAuthService(kc, silentLogger())

	boom := errors.New("connection refused")
	kc.EXPECT().CreateUser(mock.Anything, mock.Anything).Return("", boom)

	_, err := svc.Register(context.Background(), dto.RegisterRequest{
		Username: "jdoe", Email: "jdoe@lms.local", Password: "x", Role: "ROLE_USER",
	})
	assert.ErrorIs(t, err, boom)
	assert.NotErrorIs(t, err, ErrConflict)
}