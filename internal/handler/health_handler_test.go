package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHealthHandler(t *testing.T) {
	r := newTestRouter()
	r.GET("/health", Health)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/health", nil))

	assert.Equal(t, http.StatusOK, w.Code)
	var resp HealthResponse
	decodeJSON(t, w, &resp)
	assert.Equal(t, "ok", resp.Status)
}
