// Package auth validates Keycloak-issued JWTs against the realm's public
// signing keys (JWKS) and exposes the claims the rest of the app cares about.
package auth

import (
	"context"
	"fmt"
	"slices"
	"sync"

	"github.com/MicahParks/keyfunc/v3"
	"github.com/golang-jwt/jwt/v5"
)

// Claims is the subset of a Keycloak access token we rely on. Roles live in
// realm_access.roles (realm-level roles assigned to the user).
type Claims struct {
	jwt.RegisteredClaims
	PreferredUsername string `json:"preferred_username"`
	Email             string `json:"email"`
	RealmAccess       struct {
		Roles []string `json:"roles"`
	} `json:"realm_access"`
}

// HasRole reports whether the token carries the given realm role.
func (c *Claims) HasRole(role string) bool {
	return slices.Contains(c.RealmAccess.Roles, role)
}

// Validator verifies token signatures using the realm's JWKS and checks the
// issuer. The JWKS is fetched lazily on the first token so the app can start
// before Keycloak (or its realm) is ready; public routes stay available
// meanwhile. Once loaded, keyfunc refreshes the keys in the background.
type Validator struct {
	certsURL string
	issuer   string

	mu sync.Mutex
	kf keyfunc.Keyfunc
}

// NewValidator returns a Validator for the given realm. No network call is
// made here — the JWKS loads on first use.
func NewValidator(certsURL, issuer string) *Validator {
	return &Validator{certsURL: certsURL, issuer: issuer}
}

// keyfunc returns the lazily-loaded JWKS key function, retrying the initial
// fetch on each call until it succeeds (e.g. once Keycloak is reachable).
func (v *Validator) keyfunc() (keyfunc.Keyfunc, error) {
	v.mu.Lock()
	defer v.mu.Unlock()
	if v.kf != nil {
		return v.kf, nil
	}
	kf, err := keyfunc.NewDefaultCtx(context.Background(), []string{v.certsURL})
	if err != nil {
		return nil, fmt.Errorf("load JWKS from %s: %w", v.certsURL, err)
	}
	v.kf = kf
	return v.kf, nil
}

// Parse validates the signature, expiry and issuer of a raw bearer token and
// returns its claims. Any failure means the token must be rejected.
func (v *Validator) Parse(raw string) (*Claims, error) {
	kf, err := v.keyfunc()
	if err != nil {
		return nil, err
	}
	claims := &Claims{}
	_, err = jwt.ParseWithClaims(raw, claims, kf.Keyfunc,
		jwt.WithIssuer(v.issuer),
		jwt.WithValidMethods([]string{"RS256"}),
	)
	if err != nil {
		return nil, fmt.Errorf("invalid token: %w", err)
	}
	return claims, nil
}