package main

import (
	"bytes"
	"encoding/json"
	"fmt"
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
	url := filepath.Join("https://api.github.com/repos", repo.Owner, repo.Name, "hooks")
	webhook := Webhook{
		Name:   "gitfresh",
		Active: true,
		Events: []string{"push"},
		Config: map[string]string{
			"url":          config.TunnelDomain,
			"content_type": "application/json",
			"secret":       config.GitHookSecret,
		},
	}
	jsonData, err := json.Marshal(webhook)
	if err != nil {
		fmt.Println("Error al codificar la repouración del webhook:", err)
		return err
	}
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Println("Error al crear la solicitud HTTP:", err)
		return err
	}
	req.Header.Set("Authorization", "token "+config.GitServerToken)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: time.Second * 20}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error al realizar la solicitud HTTP:", err)
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		fmt.Printf("Error al crear el webhook. Código de estado: %d\n", resp.StatusCode)
		return err
	}
	fmt.Println("Webhook creado con éxito.")
	return nil
}
