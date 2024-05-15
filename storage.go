package gitfresh

import (
	"errors"
	"log/slog"
	"os"
	"path/filepath"
)

func ListRepository() ([]byte, error) {
	dir, err := os.UserHomeDir()
	if err != nil {
		return []byte{}, err
	}
	dir = filepath.Join(dir, APP_FOLDER)
	flatfile := &FlatFile{
		Name: APP_REPOS_FILE_NAME,
		Path: dir,
	}
	repos, err := flatfile.Read()
	if err != nil {
		return []byte{}, err
	}
	return repos, nil
}

type FlatFiler interface {
	Write(data []byte) (n int, err error)
	Read() (n []byte, err error)
}

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
	return len(data), nil
}

func (f *FlatFile) Read() (n []byte, err error) {
	path := filepath.Join(f.Path, f.Name)
	if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
		slog.Info("creating directory")
		_ = os.Mkdir(path, os.ModePerm)
	}
	file, err := os.ReadFile(path)
	if err != nil {
		slog.Error(err.Error())
		return []byte{}, err
	}
	return file, nil
}
