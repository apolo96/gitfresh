package gitfresh

import (
	"encoding/json"
	"errors"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func CreateConfigFile(config *AppConfig) (file string, err error) {
	dirname, err := os.UserHomeDir()
	if err != nil {
		println("error getting user home directory")
		slog.Error(err.Error())
		return file, err
	}
	content, err := json.MarshalIndent(config, "", "  ")
	slog.Debug("parsing config parameters", "data", string(content))
	if err != nil {
		println("error parsing the config parameters")
		slog.Error(err.Error())
		return file, err
	}
	path := filepath.Join(dirname, APP_FOLDER)
	if err := os.MkdirAll(path, os.ModePerm); err != nil {
		println("error making config folder")
		slog.Error(err.Error())
		return file, err
	}

	file = filepath.Join(path, APP_CONFIG_FILE_NAME)
	err = os.WriteFile(file, content, 0644)
	if err != nil {
		println("error creating app config file")
		slog.Error(err.Error())
		return file, err
	}
	slog.Info("config file created successfully")
	return file, nil
}

func ReadConfigFile() (*AppConfig, error) {
	dirname, err := os.UserHomeDir()
	config := &AppConfig{}
	if err != nil {
		return config, err
	}
	path := filepath.Join(dirname, APP_FOLDER)
	if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
		_ = os.Mkdir(path, os.ModePerm)
	}
	file, err := os.ReadFile(filepath.Join(path, APP_CONFIG_FILE_NAME))
	if err != nil {
		return config, err
	}
	if err := json.Unmarshal(file, &config); err != nil {
		return config, err
	}
	return config, nil
}

func ScanRepositories(workdir string, gitProvider string) ([]*Repository, error) {
	repos := []*Repository{}
	dirs, err := os.ReadDir(workdir)
	if err != nil {
		slog.Error(err.Error())
		return repos, err
	}
	for _, f := range dirs {
		if f.IsDir() {
			path := filepath.Join(workdir, f.Name())
			git, _ := exec.LookPath("git")
			c := exec.Command(git, "remote", "get-url", "origin")
			c.Dir = path
			url, err := c.CombinedOutput()
			if err != nil {
				slog.Info(path)
				slog.Error("executing git command", "error", err.Error(), "path", git)
				continue
			}
			slog.Info("repository remote url " + string(url))
			surl := strings.Split(string(url), "/")
			if len(surl) < 4 {
				err = errors.New("getting repository info")
				slog.Error(err.Error())
				continue
			}
			if gitProvider != surl[2] {
				err = errors.New("privider not suported" + surl[2])
				slog.Error(err.Error())
				continue
			}
			repos = append(repos, &Repository{Owner: surl[3], Name: strings.ReplaceAll(surl[4], ".git\n", "")})
		}
	}
	return repos, nil
}

func SaveReposMetaData(repos []*Repository) (file string, err error) {
	dirname, err := os.UserHomeDir()
	if err != nil {
		println("error getting user home directory")
		slog.Error(err.Error())
		return file, err
	}
	content, err := json.MarshalIndent(repos, "", "  ")
	slog.Debug("parsing config parameters", "data", string(content))
	if err != nil {
		println("error parsing the config parameters")
		slog.Error(err.Error())
		return file, err
	}
	path := filepath.Join(dirname, APP_FOLDER)
	if err := os.MkdirAll(path, os.ModePerm); err != nil {
		println("error making config folder")
		slog.Error(err.Error())
		return file, err
	}

	file = filepath.Join(path, APP_REPOS_FILE_NAME)
	err = os.WriteFile(file, content, 0644)
	if err != nil {
		println("error creating app config file")
		slog.Error(err.Error())
		return file, err
	}
	return file, nil
}

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
