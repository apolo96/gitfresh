package gitfresh

import (
	"io"
	"log/slog"
	"os"
	"path/filepath"
)

func NewLogger() (*slog.Logger, func(), error) {
	dir, _ := os.UserHomeDir()
	path := filepath.Join(dir, APP_FOLDER)
	if err := os.MkdirAll(path, os.ModePerm); err != nil {
		slog.Error(err.Error())
		return nil, func() {}, err
	}
	path = filepath.Join(path, APP_AGENT_LOG_FILE)
	logfile, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		slog.Error(err.Error())
		return nil, func() {}, err
	}
	closer := func() {
		logfile.Close()
	}
	writer := io.Writer(logfile)
	logger := slog.New(slog.NewJSONHandler(writer, &slog.HandlerOptions{Level: slog.LevelDebug}))
	logger = logger.With("version", "1.0.0")
	return logger, closer, nil
}
