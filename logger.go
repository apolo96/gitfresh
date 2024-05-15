package gitfresh

import (
	"io"
	"log/slog"
	"os"
	"path/filepath"
)


func NewLogFile(name string) (io.Writer, func(), error) {
	dir, _ := os.UserHomeDir()
	path := filepath.Join(dir, APP_FOLDER)
	if err := os.MkdirAll(path, os.ModePerm); err != nil {
		slog.Error(err.Error())
		return nil, func() {}, err
	}
	path = filepath.Join(path, name)
	logfile, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		slog.Error(err.Error())
		return nil, func() {}, err
	}
	closer := func() {
		logfile.Close()
	}
	return logfile, closer, nil
}
