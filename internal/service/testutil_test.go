package service

import (
	"io"

	"github.com/sirupsen/logrus"
)

// silentLogger returns a logrus.Logger that discards output — keeps test
// runs clean while still satisfying the *logrus.Logger dependency.
func silentLogger() *logrus.Logger {
	l := logrus.New()
	l.SetOutput(io.Discard)
	return l
}
