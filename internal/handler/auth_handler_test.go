package handler

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/dias-web/lms-system/internal/dto"
	"github.com/dias-web/lms-system/internal/service"
	svcmocks "github.com/dias-web/lms-system/internal/service/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func setupAuthRouter(t *testing.T) (*svcmocks.MockAuthService, http.Handler) {
	svc := svcmocks.NewMockAuthService(t)
	r := newTestRouter()
	NewAuthHandler(svc).Register(r)
	return svc, r
}

func TestAuthHandler_Login_OK(t *testing.T) {
	svc, r := setupAuthRouter(t)
	svc.EXPECT().
		Login(mock.Anything, dto.LoginRequest{Username: "admin", Password: "admin123"}).
		Return(dto.TokenResponse{
			AccessToken: "a.jwt", RefreshToken: "r.jwt",
			ExpiresIn: 300, RefreshExpiresIn: 604800, TokenType: "Bearer",
		}, nil)

	body := bytes.NewBufferString(`{"username":"admin","password":"admin123"}`)
	req := httptest.NewRequest(http.MethodPost, "/auth/login", body)
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp dto.TokenResponse
	decodeJSON(t, w, &resp)
	assert.Equal(t, "a.jwt", resp.AccessToken)
	assert.Equal(t, "r.jwt", resp.RefreshToken)
}

func TestAuthHandler_Login_InvalidCredentials(t *testing.T) {
	svc, r := setupAuthRouter(t)
	svc.EXPECT().Login(mock.Anything, mock.Anything).
		Return(dto.TokenResponse{}, fmt.Errorf("%w: invalid username or password", service.ErrUnauthorized))

	body := bytes.NewBufferString(`{"username":"admin","password":"wrong"}`)
	req := httptest.NewRequest(http.MethodPost, "/auth/login", body)
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	var resp dto.ErrorResponse
	decodeJSON(t, w, &resp)
	assert.Equal(t, "UNAUTHORIZED", resp.Error.Code)
}

func TestAuthHandler_Login_ValidationError(t *testing.T) {
	_, r := setupAuthRouter(t)

	// missing password
	body := bytes.NewBufferString(`{"username":"admin"}`)
	req := httptest.NewRequest(http.MethodPost, "/auth/login", body)
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	var resp dto.ErrorResponse
	decodeJSON(t, w, &resp)
	assert.Equal(t, "INVALID_INPUT", resp.Error.Code)
}

func TestAuthHandler_Refresh_OK(t *testing.T) {
	svc, r := setupAuthRouter(t)
	svc.EXPECT().
		Refresh(mock.Anything, dto.RefreshRequest{RefreshToken: "r.jwt"}).
		Return(dto.TokenResponse{
			AccessToken: "new.access", RefreshToken: "new.refresh",
			ExpiresIn: 300, RefreshExpiresIn: 604800, TokenType: "Bearer",
		}, nil)

	body := bytes.NewBufferString(`{"refresh_token":"r.jwt"}`)
	req := httptest.NewRequest(http.MethodPost, "/auth/refresh", body)
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp dto.TokenResponse
	decodeJSON(t, w, &resp)
	assert.Equal(t, "new.access", resp.AccessToken)
}

func TestAuthHandler_Refresh_Expired(t *testing.T) {
	svc, r := setupAuthRouter(t)
	svc.EXPECT().Refresh(mock.Anything, mock.Anything).
		Return(dto.TokenResponse{}, fmt.Errorf("%w: invalid or expired refresh token", service.ErrUnauthorized))

	body := bytes.NewBufferString(`{"refresh_token":"expired"}`)
	req := httptest.NewRequest(http.MethodPost, "/auth/refresh", body)
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	var resp dto.ErrorResponse
	decodeJSON(t, w, &resp)
	assert.Equal(t, "UNAUTHORIZED", resp.Error.Code)
}

func TestAuthHandler_Refresh_ValidationError(t *testing.T) {
	_, r := setupAuthRouter(t)

	body := bytes.NewBufferString(`{}`) // missing refresh_token
	req := httptest.NewRequest(http.MethodPost, "/auth/refresh", body)
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}