package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"path/filepath"
	"time"
)

/* Http Client */
type Webhook struct {
	Name   string            `json:"name"`
	Active bool              `json:"active"`
	Events []string          `json:"events"`
	Config map[string]string `json:"config"`
}

type Repository struct {
	Owner string
	Name  string
}

func createGitServerHook(repo *Repository, config *AppConfig) error {
	url := "https://api.github.com/repos/" + filepath.Join(repo.Owner, repo.Name, "hooks")
	webhook := Webhook{
		Name:   "web",
		Active: true,
		Events: []string{"push"},
		Config: map[string]string{
			"url":          config.TunnelDomain,
			"content_type": "json",
			"secret":       config.GitHookSecret,
			"insecure_ssl": "0",
		},
	}
	jsonData, err := json.Marshal(webhook)
	if err != nil {
		slog.Error(err.Error())
		return err
	}
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		slog.Error(err.Error())
		return err
	}
	req.Header.Set("Authorization", "Bearer "+config.GitServerToken)
	req.Header.Set("Accept", "*/*")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")

	client := &http.Client{Timeout: time.Second * 20}
	resp, err := client.Do(req)
	if err != nil {
		slog.Error(err.Error())
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		slog.Info(string(jsonData))
		slog.Info(url)
		b, _ := io.ReadAll(resp.Body)
		slog.Info(string(b))
		return errors.New("creating webhook via http, response with " + resp.Status)
	}
	return nil
}
