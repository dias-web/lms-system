package handler

import (
	"encoding/json"
	"io"
	"net/http/httptest"
	"testing"

	"github.com/dias-web/lms-system/internal/middleware"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

func init() { gin.SetMode(gin.TestMode) }

func silentLogger() *logrus.Logger {
	l := logrus.New()
	l.SetOutput(io.Discard)
	return l
}

// newTestRouter returns a Gin engine wired with the same middleware stack as
// production (Recovery + ErrorHandler + NoRoute/NoMethod) so handler tests
// exercise the real error-to-status mapping.
func newTestRouter() *gin.Engine {
	log := silentLogger()
	r := gin.New()
	r.HandleMethodNotAllowed = true
	r.Use(middleware.Recovery(log), middleware.ErrorHandler(log))
	r.NoRoute(middleware.NotFoundHandler())
	r.NoMethod(middleware.MethodNotAllowedHandler())
	return r
}

// decodeJSON unmarshals the recorder body into v, failing the test on error.
func decodeJSON(t *testing.T, w *httptest.ResponseRecorder, v any) {
	t.Helper()
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), v))
}
