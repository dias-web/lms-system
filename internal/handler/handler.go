package handler

import (
	"fmt"

	"github.com/dias-web/lms-system/internal/service"
)

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
