// Package keycloak wraps the gocloak client with the few operations the LMS
// backend needs: user login, token refresh and (later) admin user management.
// It returns neutral structs so the rest of the app does not depend on gocloak
// types and can be unit-tested behind an interface.
package keycloak

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/Nerzal/gocloak/v13"
	"github.com/dias-web/lms-system/internal/config"
)

// ErrInvalidCredentials is returned when Keycloak rejects the supplied
// username/password or refresh token (HTTP 400/401), so callers can map it to
// a 401 without depending on gocloak error types.
var ErrInvalidCredentials = errors.New("invalid credentials")

// ErrUserExists is returned when creating a user whose username or email is
// already taken (HTTP 409).
var ErrUserExists = errors.New("user already exists")

// ErrUserNotFound is returned when an admin operation targets a missing user.
var ErrUserNotFound = errors.New("user not found")

// Token mirrors the relevant fields of a Keycloak token response.
type Token struct {
	AccessToken      string
	RefreshToken     string
	ExpiresIn        int
	RefreshExpiresIn int
	TokenType        string
}

// adminRealm is the realm whose admin account manages users in the LMS realm.
const adminRealm = "master"

// Client talks to a single Keycloak realm using the backend's confidential
// client credentials for token flows and the master admin for user management.
type Client struct {
	gc            *gocloak.GoCloak
	realm         string
	clientID      string
	clientSecret  string
	adminUser     string
	adminPassword string
}

// NewClient builds a Keycloak client for the configured realm.
func NewClient(cfg config.KeycloakConfig) *Client {
	return &Client{
		gc:            gocloak.NewClient(cfg.URL),
		realm:         cfg.Realm,
		clientID:      cfg.ClientID,
		clientSecret:  cfg.ClientSecret,
		adminUser:     cfg.AdminUser,
		adminPassword: cfg.AdminPassword,
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

// CreateUserInput is the data needed to provision a new user with one realm
// role.
type CreateUserInput struct {
	Username  string
	Email     string
	FirstName string
	LastName  string
	Password  string
	Role      string
}

// CreateUser provisions a new enabled user with a permanent password and
// assigns the given realm role. Returns the new user's ID.
func (c *Client) CreateUser(ctx context.Context, in CreateUserInput) (string, error) {
	token, err := c.adminToken(ctx)
	if err != nil {
		return "", err
	}

	user := gocloak.User{
		Username:      gocloak.StringP(in.Username),
		Email:         gocloak.StringP(in.Email),
		FirstName:     gocloak.StringP(in.FirstName),
		LastName:      gocloak.StringP(in.LastName),
		Enabled:       gocloak.BoolP(true),
		EmailVerified: gocloak.BoolP(true),
		Credentials: &[]gocloak.CredentialRepresentation{{
			Type:      gocloak.StringP("password"),
			Value:     gocloak.StringP(in.Password),
			Temporary: gocloak.BoolP(false),
		}},
	}

	userID, err := c.gc.CreateUser(ctx, token, c.realm, user)
	if err != nil {
		var apiErr *gocloak.APIError
		if errors.As(err, &apiErr) && apiErr.Code == http.StatusConflict {
			return "", ErrUserExists
		}
		return "", err
	}

	if err := c.assignRealmRole(ctx, token, userID, in.Role); err != nil {
		return userID, err
	}
	return userID, nil
}

// User is a neutral view of a Keycloak user with its primary realm role.
type User struct {
	ID        string
	Username  string
	Email     string
	FirstName string
	LastName  string
	Role      string
}

// UpdateUserInput carries the profile fields a user may change. Empty fields
// are left untouched. Roles are intentionally absent — only admins assign them.
type UpdateUserInput struct {
	Email     string
	FirstName string
	LastName  string
}

// GetUser returns a user together with its first realm role (ROLE_*).
func (c *Client) GetUser(ctx context.Context, userID string) (*User, error) {
	token, err := c.adminToken(ctx)
	if err != nil {
		return nil, err
	}
	u, err := c.gc.GetUserByID(ctx, token, c.realm, userID)
	if err != nil {
		return nil, mapUserError(err)
	}
	role, err := c.primaryRealmRole(ctx, token, userID)
	if err != nil {
		return nil, err
	}
	return &User{
		ID:        deref(u.ID),
		Username:  deref(u.Username),
		Email:     deref(u.Email),
		FirstName: deref(u.FirstName),
		LastName:  deref(u.LastName),
		Role:      role,
	}, nil
}

// UpdateUser applies the non-empty profile fields to an existing user. It reads
// the current representation first so unspecified fields are preserved.
func (c *Client) UpdateUser(ctx context.Context, userID string, in UpdateUserInput) error {
	token, err := c.adminToken(ctx)
	if err != nil {
		return err
	}
	u, err := c.gc.GetUserByID(ctx, token, c.realm, userID)
	if err != nil {
		return mapUserError(err)
	}
	if in.Email != "" {
		u.Email = gocloak.StringP(in.Email)
	}
	if in.FirstName != "" {
		u.FirstName = gocloak.StringP(in.FirstName)
	}
	if in.LastName != "" {
		u.LastName = gocloak.StringP(in.LastName)
	}
	if err := c.gc.UpdateUser(ctx, token, c.realm, *u); err != nil {
		return mapUserError(err)
	}
	return nil
}

// SetPassword sets a new permanent password for the user.
func (c *Client) SetPassword(ctx context.Context, userID, newPassword string) error {
	token, err := c.adminToken(ctx)
	if err != nil {
		return err
	}
	if err := c.gc.SetPassword(ctx, token, userID, c.realm, newPassword, false); err != nil {
		return mapUserError(err)
	}
	return nil
}

func (c *Client) primaryRealmRole(ctx context.Context, token, userID string) (string, error) {
	roles, err := c.gc.GetRealmRolesByUserID(ctx, token, c.realm, userID)
	if err != nil {
		return "", err
	}
	for _, r := range roles {
		if name := deref(r.Name); strings.HasPrefix(name, "ROLE_") {
			return name, nil
		}
	}
	return "", nil
}

func (c *Client) assignRealmRole(ctx context.Context, token, userID, roleName string) error {
	role, err := c.gc.GetRealmRole(ctx, token, c.realm, roleName)
	if err != nil {
		return err
	}
	return c.gc.AddRealmRoleToUser(ctx, token, c.realm, userID, []gocloak.Role{*role})
}

// adminToken logs into the master realm and returns an admin access token.
func (c *Client) adminToken(ctx context.Context) (string, error) {
	jwt, err := c.gc.LoginAdmin(ctx, c.adminUser, c.adminPassword, adminRealm)
	if err != nil {
		return "", err
	}
	return jwt.AccessToken, nil
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

// mapUserError maps admin user-management errors: 404 -> ErrUserNotFound,
// 409 -> ErrUserExists (e.g. email already taken), everything else unchanged.
func mapUserError(err error) error {
	var apiErr *gocloak.APIError
	if errors.As(err, &apiErr) {
		switch apiErr.Code {
		case http.StatusNotFound:
			return ErrUserNotFound
		case http.StatusConflict:
			return ErrUserExists
		}
	}
	return err
}

func deref(s *string) string {
	if s == nil {
		return ""
	}
	return *s
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