// Package keycloak wraps the gocloak client with the few operations the LMS
// backend needs: user login, token refresh and (later) admin user management.
// It returns neutral structs so the rest of the app does not depend on gocloak
// types and can be unit-tested behind an interface.
package keycloak

import (
	"context"
	"errors"
	"net/http"

	"github.com/Nerzal/gocloak/v13"
	"github.com/dias-web/lms-system/internal/config"
)

// ErrInvalidCredentials is returned when Keycloak rejects the supplied
// username/password or refresh token (HTTP 400/401), so callers can map it to
// a 401 without depending on gocloak error types.
var ErrInvalidCredentials = errors.New("invalid credentials")

// Token mirrors the relevant fields of a Keycloak token response.
type Token struct {
	AccessToken      string
	RefreshToken     string
	ExpiresIn        int
	RefreshExpiresIn int
	TokenType        string
}

// Client talks to a single Keycloak realm using the backend's confidential
// client credentials.
type Client struct {
	gc           *gocloak.GoCloak
	realm        string
	clientID     string
	clientSecret string
}

// NewClient builds a Keycloak client for the configured realm.
func NewClient(cfg config.KeycloakConfig) *Client {
	return &Client{
		gc:           gocloak.NewClient(cfg.URL),
		realm:        cfg.Realm,
		clientID:     cfg.ClientID,
		clientSecret: cfg.ClientSecret,
	}
}

// Login performs the OAuth2 password grant and returns the issued tokens.
func (c *Client) Login(ctx context.Context, username, password string) (*Token, error) {
	jwt, err := c.gc.Login(ctx, c.clientID, c.clientSecret, c.realm, username, password)
	if err != nil {
		return nil, translateError(err)
	}
	return toToken(jwt), nil
}

// Refresh exchanges a refresh token for a fresh pair of tokens.
func (c *Client) Refresh(ctx context.Context, refreshToken string) (*Token, error) {
	jwt, err := c.gc.RefreshToken(ctx, refreshToken, c.clientID, c.clientSecret, c.realm)
	if err != nil {
		return nil, translateError(err)
	}
	return toToken(jwt), nil
}

// translateError maps Keycloak's 400/401 rejections to ErrInvalidCredentials
// and leaves everything else (network, 5xx) as-is.
func translateError(err error) error {
	var apiErr *gocloak.APIError
	if errors.As(err, &apiErr) &&
		(apiErr.Code == http.StatusUnauthorized || apiErr.Code == http.StatusBadRequest) {
		return ErrInvalidCredentials
	}
	return err
}

func toToken(jwt *gocloak.JWT) *Token {
	return &Token{
		AccessToken:      jwt.AccessToken,
		RefreshToken:     jwt.RefreshToken,
		ExpiresIn:        jwt.ExpiresIn,
		RefreshExpiresIn: jwt.RefreshExpiresIn,
		TokenType:        jwt.TokenType,
	}
}