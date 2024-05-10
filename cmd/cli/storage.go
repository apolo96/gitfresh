package main

import (
	"encoding/json"
	"errors"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

/* FlatFile */
const APP_CONFIG_FILE_NAME = "config.json"
const APP_CONFIG_FOLDER = ".gitfresh"
const APP_REPOS_FILE_NAME = "repositories.json"

type AppConfig struct {
	TunnelToken    string
	TunnelDomain   string
	GitServerToken string
	GitWorkDir     string
	GitHookSecret  string
}

func createConfigFile(config *AppFlags) (file string, err error) {
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
	path := filepath.Join(dirname, APP_CONFIG_FOLDER)
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

func readConfigFile() (*AppConfig, error) {
	dirname, err := os.UserHomeDir()
	config := &AppConfig{}
	if err != nil {
		return config, err
	}
	path := filepath.Join(dirname, APP_CONFIG_FOLDER)
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

func scanRepositories(workdir string, gitProvider string) ([]Repository, error) {
	repos := []Repository{}
	files, err := os.ReadDir(workdir)
	if err != nil {
		slog.Error(err.Error())
		return repos, err
	}
	for _, f := range files {
		if f.IsDir() {
			path := filepath.Join(workdir, f.Name())
			err := os.Chdir(path)
			if err != nil {
				break
			}
			git, _ := exec.LookPath("git")
			url, err := exec.Command(git, "remote", "get-url", "origin").CombinedOutput()
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
			repos = append(repos, Repository{Owner: surl[3], Name: strings.ReplaceAll(surl[4], ".git\n", "")})
			os.Chdir(workdir)
		}
	}
	return repos, nil
}

func saveReposMetaData(repos []*Repository) (file string, err error) {
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
	path := filepath.Join(dirname, APP_CONFIG_FOLDER)
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
