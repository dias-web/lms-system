package middleware

import (
	"fmt"
	"strings"

	"github.com/dias-web/lms-system/internal/auth"
	"github.com/dias-web/lms-system/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// Context keys under which the authenticated identity is stored.
const (
	ctxClaims   = "auth.claims"
	ctxUsername = "auth.username"
	ctxRoles    = "auth.roles"
)

// AuthRequired validates the Bearer token on the request. On success the
// token claims are stored in the gin context for downstream handlers and
// RequireRole. On failure it attaches a 401 error and aborts the chain; the
// global ErrorHandler turns that into the unified JSON response.
func AuthRequired(v *auth.Validator, log *logrus.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		raw, err := bearerToken(c)
		if err != nil {
			abortUnauthorized(c, err)
			return
		}

		claims, err := v.Parse(raw)
		if err != nil {
			log.WithError(err).
				WithField("path", c.Request.URL.Path).
				Debug("token validation failed")
			abortUnauthorized(c, fmt.Errorf("%w: %s", service.ErrUnauthorized, "invalid or expired token"))
			return
		}

		c.Set(ctxClaims, claims)
		c.Set(ctxUsername, claims.PreferredUsername)
		c.Set(ctxRoles, claims.RealmAccess.Roles)
		c.Next()
	}
}

// RequireRole ensures the authenticated user carries at least one of the
// given realm roles. Must run after AuthRequired.
func RequireRole(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		claims, ok := CurrentClaims(c)
		if !ok {
			abortUnauthorized(c, fmt.Errorf("%w: %s", service.ErrUnauthorized, "missing authentication"))
			return
		}
		for _, role := range roles {
			if claims.HasRole(role) {
				c.Next()
				return
			}
		}
		_ = c.Error(fmt.Errorf("%w: requires one of roles %v", service.ErrForbidden, roles))
		c.Abort()
	}
}

// CurrentClaims returns the authenticated user's claims, if any.
func CurrentClaims(c *gin.Context) (*auth.Claims, bool) {
	v, ok := c.Get(ctxClaims)
	if !ok {
		return nil, false
	}
	claims, ok := v.(*auth.Claims)
	return claims, ok
}

func bearerToken(c *gin.Context) (string, error) {
	h := c.GetHeader("Authorization")
	if h == "" {
		return "", fmt.Errorf("%w: %s", service.ErrUnauthorized, "missing Authorization header")
	}
	parts := strings.SplitN(h, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") || parts[1] == "" {
		return "", fmt.Errorf("%w: %s", service.ErrUnauthorized, "malformed Authorization header")
	}
	return parts[1], nil
}

func abortUnauthorized(c *gin.Context, err error) {
	_ = c.Error(err)
	c.Abort()
}