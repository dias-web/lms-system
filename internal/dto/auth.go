package dto

// LoginRequest is the body for POST /auth/login.
type LoginRequest struct {
	Username string `json:"username" binding:"required" example:"admin"`
	Password string `json:"password" binding:"required" example:"admin123"`
} // @name LoginRequest

// TokenResponse is the token pair returned by login and refresh endpoints.
// It mirrors the relevant fields of a Keycloak token response.
type TokenResponse struct {
	AccessToken      string `json:"access_token"`
	RefreshToken     string `json:"refresh_token"`
	ExpiresIn        int    `json:"expires_in" example:"300"`
	RefreshExpiresIn int    `json:"refresh_expires_in" example:"604800"`
	TokenType        string `json:"token_type" example:"Bearer"`
} // @name TokenResponse