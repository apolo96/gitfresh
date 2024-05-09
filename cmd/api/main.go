package main

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"os"
	"os/exec"
	"strings"

	"golang.ngrok.com/ngrok"
	"golang.ngrok.com/ngrok/config"
)

func main() {
	if err := run(context.Background()); err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}
}

func run(ctx context.Context) error {
	secret := os.Getenv("GITHUB_HOOK_SECRET")
	slog.Debug("git provider webhook secret", "secret", secret)
	listener, err := ngrok.Listen(ctx,
		config.HTTPEndpoint(
			config.WithWebhookVerification("github", secret),
			config.WithDomain("yak-loyal-violently.ngrok-free.app"),
		),
		ngrok.WithAuthtokenFromEnv(),
	)
	if err != nil {
		return err
	}
	slog.Info("Listening on " + listener.URL())
	return http.Serve(listener, http.HandlerFunc(handler))
}

type Repository struct {
	Name string `json:"name"`
}

type Payload struct {
	Ref        string     `json:"ref"`
	Repository Repository `json:"repository"`
	Commit     string     `json:"after"`
}

func handler(w http.ResponseWriter, r *http.Request) {
	slog.Info("handling webhook")
	slog.Info("parse http form body")
	err := r.ParseForm()
	if err != nil {
		slog.Error("parsing http data form", "error", err.Error())
		http.Error(w, "error parsing data form", http.StatusBadRequest)
		return
	}
	form := r.FormValue("payload")
	var p Payload
	if err := json.Unmarshal([]byte(form), &p); err != nil {
		slog.Error(err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	slog.Info(
		"payload form data",
		"branch", p.Ref,
		"repository", p.Repository.Name,
		"last_commit", p.Commit,
	)
	gitPullCmd("/Users/laniakea/code/", p.Repository.Name, p.Ref)
	w.WriteHeader(http.StatusOK)
}

func gitPullCmd(workdir, repoName, branch string) (err error) {
	p, err := exec.Command("which", "git").CombinedOutput()
	if err != nil {
		slog.Error("which git path", "error", err.Error())
		return err
	}
	path := strings.ReplaceAll(string(p), "\n", "")
	slog.Info("which command output", "path", path)
	workspace := workdir + repoName
	cmd := exec.Command(path, "pull", "origin", branch)
	cmd.Dir = workspace
	slog.Info("running command ", "cmd", cmd.String(), "dir", cmd.Dir)
	err = cmd.Run()
	if err != nil {
		slog.Error("git command failed", "error", err.Error())
	}
	return err
}
