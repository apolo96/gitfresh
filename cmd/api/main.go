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
	if err := run(); err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}
}

func run() (e error) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	ch := make(chan string)
	go func() {
		if err := tunnel(ctx, ch); err != nil {
			cancel()
			slog.Error(err.Error())
			e = err
		}
	}()
	go func() {
		if err := localserver(ctx, ch); err != nil {
			cancel()
			slog.Error(err.Error())
			e = err
		}
	}()
	<-ctx.Done()
	println("Stop Service by error cause")
	return e
}

func localserver(ctx context.Context, ch <-chan string) error {
	url := <-ch
	server := &http.Server{
		Addr: "127.0.0.1:9191",
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("Service StatusOK On " + url))
		}),
	}
	slog.Info("LocalServer Listening on " + server.Addr)
	return server.ListenAndServe()
}

func tunnel(ctx context.Context, ch chan<- string) error {
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
	ch <- listener.URL()
	slog.Info("Tunnel Listening on " + listener.URL())
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
		"last_commit", p.Commit[:7],
	)
	gitPullCmd("/Users/laniakea/code/", p.Repository.Name, p.Ref)
	w.WriteHeader(http.StatusOK)
}

func gitPullCmd(workdir, repoName, branch string) (err error) {
	p, err := exec.LookPath("git")
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
