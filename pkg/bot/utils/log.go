package utils

import (
	"io"
	"os"
	"path"
	"time"

	"github.com/rs/zerolog"
)

func NewBotLogger(dev bool, logPath string) (zerolog.Logger, error) {
	output, err := getLoggerOuput(dev, logPath)
	if err != nil {
		return zerolog.New(os.Stderr).With().Timestamp().Str("part", "bot").Logger(), err
	}
	return zerolog.New(output).With().Timestamp().Str("part", "bot").Logger(), nil
}

func NewSourceDeactiverLogger(dev bool, logPath string) (zerolog.Logger, error) {
	output, err := getLoggerOuput(dev, logPath)
	if err != nil {
		return zerolog.New(os.Stderr).With().Timestamp().Str("part", "deactiver").Logger(), err
	}
	return zerolog.New(output).With().Timestamp().Str("part", "bot").Logger(), nil
}

func getLoggerOuput(dev bool, logPath string) (io.Writer, error) {
	if dev {
		return zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.RFC3339}, nil
	}

	logDir := path.Dir(logPath)
	// TODO use a better permission for log path
	if err := os.MkdirAll(logDir, 0777); err != nil {
		return nil, err
	}

	return os.Open(logPath)
}
