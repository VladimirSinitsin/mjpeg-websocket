package logger

import (
	"os"
	"strings"

	"github.com/rs/zerolog"
)

func New(level string) *zerolog.Logger {
	var l zerolog.Level

	switch strings.ToLower(level) {
	case zerolog.ErrorLevel.String():
		l = zerolog.ErrorLevel
	case zerolog.WarnLevel.String():
		l = zerolog.WarnLevel
	case zerolog.InfoLevel.String():
		l = zerolog.InfoLevel
	case zerolog.DebugLevel.String():
		l = zerolog.DebugLevel
	default:
		l = zerolog.InfoLevel
	}

	zerolog.SetGlobalLevel(l)

	logger := zerolog.New(os.Stdout).With().Timestamp().Logger()

	return &logger
}
