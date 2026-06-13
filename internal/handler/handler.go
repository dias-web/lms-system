package handler

import (
	"fmt"

	"github.com/dias-web/lms-system/internal/service"
	"github.com/gin-gonic/gin"
)

// chain prepends route-level middleware (e.g. auth) to a final handler,
// returning the slice gin expects. A fresh slice is allocated each call so
// repeated use across routes cannot alias a shared backing array.
func chain(mw []gin.HandlerFunc, final gin.HandlerFunc) []gin.HandlerFunc {
	out := make([]gin.HandlerFunc, 0, len(mw)+1)
	out = append(out, mw...)
	return append(out, final)
}

// parseUintParam parses a positive integer path parameter.
func parseUintParam(s string, dst *uint) error {
	if _, err := fmt.Sscanf(s, "%d", dst); err != nil {
		return err
	}
	return nil
}

// bindError wraps a Gin binding error so the ErrorHandler middleware emits
// 400 INVALID_INPUT while preserving the validator's message.
func bindError(err error) error {
	return fmt.Errorf("%w: %s", service.ErrInvalidInput, err.Error())
}

// invalidIDError signals an unparsable :id path parameter.
func invalidIDError() error {
	return fmt.Errorf("%w: invalid id", service.ErrInvalidInput)
}
