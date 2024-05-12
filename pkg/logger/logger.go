package logger

import (
	"os"

	"github.com/rs/zerolog"
)

func New(lvl string) (zerolog.Logger, error) {
	logLevel, err := zerolog.ParseLevel(lvl)
	if err != nil {
		return zerolog.Logger{}, err
	}

	logger := zerolog.New(os.Stdout).
		Level(logLevel).
		With().
		Timestamp().
		Caller().
		Logger()

	return logger, nil
}
