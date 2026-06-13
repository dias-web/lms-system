package handler

import (
	"net/http"

	"github.com/dias-web/lms-system/internal/dto"
	"github.com/dias-web/lms-system/internal/service"
	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	svc service.AuthService
}

func NewAuthHandler(svc service.AuthService) *AuthHandler {
	return &AuthHandler{svc: svc}
}

// Register wires the authentication routes. login/refresh are public; the
// adminOnly handlers (auth + ROLE_ADMIN guard) protect user registration.
func (h *AuthHandler) Register(r *gin.Engine, adminOnly ...gin.HandlerFunc) {
	r.POST("/auth/login", h.Login)
	r.POST("/auth/refresh", h.Refresh)
	r.POST("/auth/register", chain(adminOnly, h.RegisterUser)...)
}

// Login godoc
// @Summary      Authenticate and obtain tokens
// @Description  Exchanges username and password for a JWT access token (5 min)
// @Description  and a refresh token (168 h), both issued by Keycloak.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        credentials  body      dto.LoginRequest  true  "User credentials"
// @Success      200  {object}  dto.TokenResponse
// @Failure      400  {object}  dto.ErrorResponse  "Validation failed"
// @Failure      401  {object}  dto.ErrorResponse  "Invalid username or password"
// @Router       /auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var req dto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(bindError(err))
		return
	}
	tokens, err := h.svc.Login(c.Request.Context(), req)
	if err != nil {
		_ = c.Error(err)
		return
	}
	c.JSON(http.StatusOK, tokens)
}

// Refresh godoc
// @Summary      Refresh access token
// @Description  Exchanges a valid refresh token for a new access/refresh token
// @Description  pair. Used when the 5-minute access token expires.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        token  body      dto.RefreshRequest  true  "Refresh token"
// @Success      200  {object}  dto.TokenResponse
// @Failure      400  {object}  dto.ErrorResponse  "Validation failed"
// @Failure      401  {object}  dto.ErrorResponse  "Invalid or expired refresh token"
// @Router       /auth/refresh [post]
func (h *AuthHandler) Refresh(c *gin.Context) {
	var req dto.RefreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(bindError(err))
		return
	}
	tokens, err := h.svc.Refresh(c.Request.Context(), req)
	if err != nil {
		_ = c.Error(err)
		return
	}
	c.JSON(http.StatusOK, tokens)
}

// RegisterUser godoc
// @Summary      Register a new user (admin only)
// @Description  Creates a Keycloak user and assigns a realm role. Requires a
// @Description  valid token carrying ROLE_ADMIN.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        user  body      dto.RegisterRequest  true  "New user"
// @Success      201  {object}  dto.UserResponse
// @Failure      400  {object}  dto.ErrorResponse  "Validation failed"
// @Failure      401  {object}  dto.ErrorResponse  "Missing or invalid token"
// @Failure      403  {object}  dto.ErrorResponse  "Caller is not an admin"
// @Failure      409  {object}  dto.ErrorResponse  "Username or email already taken"
// @Router       /auth/register [post]
func (h *AuthHandler) RegisterUser(c *gin.Context) {
	var req dto.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(bindError(err))
		return
	}
	user, err := h.svc.Register(c.Request.Context(), req)
	if err != nil {
		_ = c.Error(err)
		return
	}
	c.JSON(http.StatusCreated, user)
}