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

// Register wires the authentication routes. These are public — they are how a
// client obtains a token in the first place.
func (h *AuthHandler) Register(r *gin.Engine) {
	r.POST("/auth/login", h.Login)
	r.POST("/auth/refresh", h.Refresh)
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