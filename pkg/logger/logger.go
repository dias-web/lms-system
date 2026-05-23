package logger

import (
	"os"

	"github.com/sirupsen/logrus"
)

// New builds a logrus logger.
//
// Formatter is chosen by env: JSON in production, human-readable text
// otherwise. Level is resolved in this order:
//  1. explicit LOG_LEVEL argument (debug/info/warn/error/fatal/panic),
//  2. fallback by env — DEBUG in non-production, INFO in production.
func New(env, level string) *logrus.Logger {
	log := logrus.New()
	log.SetOutput(os.Stdout)

	if env == "production" {
		log.SetFormatter(&logrus.JSONFormatter{})
	} else {
		log.SetFormatter(&logrus.TextFormatter{FullTimestamp: true})
	}

	if level != "" {
		if parsed, err := logrus.ParseLevel(level); err == nil {
			log.SetLevel(parsed)
			log.Debugf("log level set from LOG_LEVEL=%s", level)
			return log
		}
		log.Warnf("invalid LOG_LEVEL=%q, falling back to env default", level)
	}

	if env == "production" {
		log.SetLevel(logrus.InfoLevel)
	} else {
		log.SetLevel(logrus.DebugLevel)
	}
	return log
}
