package middleware

import (
	"net/http"

	"github.com/dias-web/lms-system/internal/dto"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// Recovery catches panics and returns a standard JSON 500 response,
// logging the panic value and request context.
func Recovery(log *logrus.Logger) gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered any) {
		log.WithFields(logrus.Fields{
			"panic":  recovered,
			"path":   c.Request.URL.Path,
			"method": c.Request.Method,
		}).Error("panic recovered")
		c.AbortWithStatusJSON(http.StatusInternalServerError, dto.NewErrorResponse(
			"INTERNAL_ERROR",
			"internal server error",
		))
	})
}
