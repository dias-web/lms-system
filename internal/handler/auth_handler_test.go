package handler

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/dias-web/lms-system/internal/auth"
	"github.com/dias-web/lms-system/internal/dto"
	"github.com/dias-web/lms-system/internal/middleware"
	"github.com/dias-web/lms-system/internal/service"
	svcmocks "github.com/dias-web/lms-system/internal/service/mocks"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// testClaims is the identity injected into the auth router for handler tests:
// a user with id "test-user-id" / username "tester".
func testClaims() *auth.Claims {
	cl := &auth.Claims{PreferredUsername: "tester"}
	cl.Subject = "test-user-id"
	cl.RealmAccess.Roles = []string{"ROLE_USER"}
	return cl
}

func setupAuthRouter(t *testing.T) (*svcmocks.MockAuthService, http.Handler) {
	svc := svcmocks.NewMockAuthService(t)
	r := newTestRouter()
	// Inject a fixed authenticated user; the admin guard is a no-op so handler
	// logic (not the role check) is what these tests exercise.
	authStub := middleware.InjectClaims(testClaims())
	adminStub := func(c *gin.Context) { c.Next() }
	NewAuthHandler(svc).Register(r, authStub, adminStub)
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

func TestAuthHandler_Register_OK(t *testing.T) {
	svc, r := setupAuthRouter(t)
	svc.EXPECT().
		Register(mock.Anything, dto.RegisterRequest{
			Username: "jdoe", Email: "jdoe@lms.local", Password: "secret123",
			FirstName: "John", LastName: "Doe", Role: "ROLE_USER",
		}).
		Return(dto.UserResponse{ID: "uuid-1", Username: "jdoe", Role: "ROLE_USER"}, nil)

	body := bytes.NewBufferString(`{"username":"jdoe","email":"jdoe@lms.local","password":"secret123","first_name":"John","last_name":"Doe","role":"ROLE_USER"}`)
	req := httptest.NewRequest(http.MethodPost, "/auth/register", body)
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
	var resp dto.UserResponse
	decodeJSON(t, w, &resp)
	assert.Equal(t, "uuid-1", resp.ID)
}

func TestAuthHandler_Register_Conflict(t *testing.T) {
	svc, r := setupAuthRouter(t)
	svc.EXPECT().Register(mock.Anything, mock.Anything).
		Return(dto.UserResponse{}, fmt.Errorf("%w: username or email already taken", service.ErrConflict))

	body := bytes.NewBufferString(`{"username":"admin","email":"admin@lms.local","password":"secret123","role":"ROLE_USER"}`)
	req := httptest.NewRequest(http.MethodPost, "/auth/register", body)
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusConflict, w.Code)
	var resp dto.ErrorResponse
	decodeJSON(t, w, &resp)
	assert.Equal(t, "CONFLICT", resp.Error.Code)
}

func TestAuthHandler_Register_InvalidRole(t *testing.T) {
	_, r := setupAuthRouter(t)

	// role not in the allowed oneof set -> validation 400 before service call
	body := bytes.NewBufferString(`{"username":"jdoe","email":"jdoe@lms.local","password":"secret123","role":"ROLE_SUPERHERO"}`)
	req := httptest.NewRequest(http.MethodPost, "/auth/register", body)
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	var resp dto.ErrorResponse
	decodeJSON(t, w, &resp)
	assert.Equal(t, "INVALID_INPUT", resp.Error.Code)
}

func TestAuthHandler_UpdateProfile_OK(t *testing.T) {
	svc, r := setupAuthRouter(t)
	// userID comes from injected claims (Subject = "test-user-id")
	svc.EXPECT().
		UpdateProfile(mock.Anything, "test-user-id", dto.UpdateProfileRequest{
			Email: "new@lms.local", FirstName: "New", LastName: "Name",
		}).
		Return(dto.UserResponse{ID: "test-user-id", Username: "tester", Email: "new@lms.local", Role: "ROLE_USER"}, nil)

	body := bytes.NewBufferString(`{"email":"new@lms.local","first_name":"New","last_name":"Name"}`)
	req := httptest.NewRequest(http.MethodPut, "/auth/profile", body)
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp dto.UserResponse
	decodeJSON(t, w, &resp)
	assert.Equal(t, "new@lms.local", resp.Email)
	assert.Equal(t, "ROLE_USER", resp.Role)
}

func TestAuthHandler_UpdateProfile_InvalidEmail(t *testing.T) {
	_, r := setupAuthRouter(t)

	body := bytes.NewBufferString(`{"email":"not-an-email"}`)
	req := httptest.NewRequest(http.MethodPut, "/auth/profile", body)
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAuthHandler_ChangePassword_OK(t *testing.T) {
	svc, r := setupAuthRouter(t)
	svc.EXPECT().
		ChangePassword(mock.Anything, "test-user-id", "tester", dto.ChangePasswordRequest{
			OldPassword: "oldpass", NewPassword: "newpass123",
		}).
		Return(nil)

	body := bytes.NewBufferString(`{"old_password":"oldpass","new_password":"newpass123"}`)
	req := httptest.NewRequest(http.MethodPut, "/auth/password", body)
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)
	assert.Empty(t, w.Body.String())
}

func TestAuthHandler_ChangePassword_WrongCurrent(t *testing.T) {
	svc, r := setupAuthRouter(t)
	svc.EXPECT().ChangePassword(mock.Anything, "test-user-id", "tester", mock.Anything).
		Return(fmt.Errorf("%w: current password is incorrect", service.ErrUnauthorized))

	body := bytes.NewBufferString(`{"old_password":"wrong","new_password":"newpass123"}`)
	req := httptest.NewRequest(http.MethodPut, "/auth/password", body)
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	var resp dto.ErrorResponse
	decodeJSON(t, w, &resp)
	assert.Equal(t, "UNAUTHORIZED", resp.Error.Code)
}

func TestAuthHandler_ChangePassword_ValidationError(t *testing.T) {
	_, r := setupAuthRouter(t)

	// new_password too short (min=6)
	body := bytes.NewBufferString(`{"old_password":"oldpass","new_password":"123"}`)
	req := httptest.NewRequest(http.MethodPut, "/auth/password", body)
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}