package gitfresh

import (
	"errors"
	"log/slog"
	"os"
	"path/filepath"
)

type FlatFile struct {
	Name string
	Path string
}

func (f *FlatFile) Write(data []byte) (n int, err error) {
	if err := os.MkdirAll(f.Path, os.ModePerm); err != nil {
		slog.Error(err.Error())
		return 0, err
	}
	err = os.WriteFile(filepath.Join(f.Path, f.Name), data, 0644)
	if err != nil {
		slog.Error(err.Error())
		return 0, err
	}
	slog.Info("config file created successfully")
	return len(p), nil
}

func (f *FlatFile) Read() (n []byte, err error) {
	path := filepath.Join(f.Path, f.Name)
	if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
		_ = os.Mkdir(path, os.ModePerm)
	}
	file, err := os.ReadFile(path)
	if err != nil {
		return []byte{}, err
	}
	return file, nil
}
