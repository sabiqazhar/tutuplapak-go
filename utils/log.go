package utils

import (
	"os"

	"github.com/rs/zerolog"
)

var Logger *zerolog.Logger

func InitLogger() {
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	logger := zerolog.New(zerolog.ConsoleWriter{Out: os.Stdout}).With().Timestamp().Logger()
	Logger = &logger
}
