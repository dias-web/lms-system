package middleware

import (
	"errors"
	"net/http"

	"github.com/dias-web/lms-system/internal/dto"
	"github.com/dias-web/lms-system/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// ErrorHandler converts errors attached via c.Error(...) into a standard
// JSON response. Handlers should call c.Error(err); return — and let the
// middleware decide the status code based on the error type.
func ErrorHandler(log *logrus.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		if len(c.Errors) == 0 {
			return
		}
		if c.Writer.Written() {
			return
		}

		err := c.Errors.Last().Err

		switch {
		case errors.Is(err, service.ErrCourseNotFound):
			c.JSON(http.StatusNotFound, dto.NewErrorResponse("COURSE_NOT_FOUND", err.Error()))
		case errors.Is(err, service.ErrChapterNotFound):
			c.JSON(http.StatusNotFound, dto.NewErrorResponse("CHAPTER_NOT_FOUND", err.Error()))
		case errors.Is(err, service.ErrLessonNotFound):
			c.JSON(http.StatusNotFound, dto.NewErrorResponse("LESSON_NOT_FOUND", err.Error()))
		case errors.Is(err, service.ErrNotFound):
			c.JSON(http.StatusNotFound, dto.NewErrorResponse("NOT_FOUND", err.Error()))
		case errors.Is(err, service.ErrInvalidInput):
			c.JSON(http.StatusBadRequest, dto.NewErrorResponse("INVALID_INPUT", err.Error()))
		case errors.Is(err, service.ErrUnauthorized):
			c.JSON(http.StatusUnauthorized, dto.NewErrorResponse("UNAUTHORIZED", err.Error()))
		case errors.Is(err, service.ErrForbidden):
			c.JSON(http.StatusForbidden, dto.NewErrorResponse("FORBIDDEN", err.Error()))
		default:
			log.WithError(err).
				WithField("path", c.Request.URL.Path).
				WithField("method", c.Request.Method).
				Error("unhandled error")
			c.JSON(http.StatusInternalServerError, dto.NewErrorResponse(
				"INTERNAL_ERROR",
				"internal server error",
			))
		}
	}
}

// NotFoundHandler returns a standard JSON 404 for unknown routes.
func NotFoundHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusNotFound, dto.NewErrorResponse(
			"ROUTE_NOT_FOUND",
			"route not found: "+c.Request.Method+" "+c.Request.URL.Path,
		))
	}
}

// MethodNotAllowedHandler returns a standard JSON 405 for unsupported methods.
func MethodNotAllowedHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusMethodNotAllowed, dto.NewErrorResponse(
			"METHOD_NOT_ALLOWED",
			"method not allowed: "+c.Request.Method+" "+c.Request.URL.Path,
		))
	}
}
