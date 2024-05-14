package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/apolo96/gitfresh"
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
	/* Logger */
	println("Config Agent Logger")
	logger, close, err := gitfresh.NewLogger()
	if err != nil {
		return err
	}
	defer close()
	slog.SetDefault(logger)
	/* Servers */
	slog.Info("Loading GitFresh Agent")
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
		if err := localserver(ch); err != nil {
			cancel()
			slog.Error(err.Error())
			e = err
		}
	}()
	<-ctx.Done()
	slog.Error("localserver or tunnel failed", "error", e.Error())
	println("localserver or tunnel failed")
	return e
}

func localserver(ch <-chan string) error {
	url := <-ch
	server := &http.Server{
		Addr: gitfresh.API_AGENT_HOST,
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			data := fmt.Sprintf(`{"api_version":"1.0.0", "tunnel_domain":"%s"}`, url)
			w.Header().Set("Content-type", "application/json")
			w.Write([]byte(data))
		}),
	}
	msg := "LocalServer Listening on " + server.Addr
	println(msg)
	slog.Info(msg)
	return server.ListenAndServe()
}

func tunnel(ctx context.Context, ch chan<- string) error {
	conf, err := gitfresh.ReadConfigFile()
	if err != nil {
		slog.Error(err.Error())
		return err
	}
	slog.Debug("load agent config from file", "config", fmt.Sprint(conf))
	os.Setenv("NGROK_AUTHTOKEN", conf.TunnelToken)
	listener, err := ngrok.Listen(ctx,
		config.HTTPEndpoint(
			config.WithWebhookVerification("github", conf.GitHookSecret),
			config.WithDomain("yak-loyal-violently.ngrok-free.app"),
		),
		ngrok.WithAuthtokenFromEnv(),
	)
	if err != nil {
		return err
	}
	ch <- listener.URL()
	msg := "Tunnel Listening on " + listener.URL()
	println(msg)
	slog.Info(msg)
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
	slog.Info("handling webhook", "host", w.Header().Get("Host"))
	form, err := io.ReadAll(r.Body)
	if err != nil {
		slog.Error(err.Error())
		http.Error(w, "error reading request data", http.StatusBadRequest)
		return
	}
	var p Payload
	if err := json.Unmarshal([]byte(form), &p); err != nil {
		slog.Error(err.Error())
		http.Error(w, "error parsing data form", http.StatusBadRequest)
		return
	}
	slog.Info(
		"payload form data",
		"branch", p.Ref,
		"repository", p.Repository.Name,
		"last_commit", p.Commit[:7],
	)
	w.WriteHeader(http.StatusOK)
	go func() {
		conf, err := gitfresh.ReadConfigFile()
		if err != nil {
			slog.Error(err.Error())
		}
		gitPullCmd(conf.GitWorkDir, p.Repository.Name, p.Ref)
	}()
}

func gitPullCmd(workdir, repoName, branch string) (err error) {
	p, err := exec.LookPath("git")
	if err != nil {
		slog.Error("which git path", "error", err.Error())
		return err
	}
	path := strings.ReplaceAll(string(p), "\n", "")
	slog.Info("which command output", "path", path)
	workspace := filepath.Join(workdir, repoName)
	cmd := exec.Command(path, "pull", "origin", branch)
	cmd.Dir = workspace
	slog.Info("running command ", "cmd", cmd.String(), "dir", cmd.Dir)
	err = cmd.Run()
	if err != nil {
		slog.Error("git command failed", "error", err.Error())
	}
	return err
}
