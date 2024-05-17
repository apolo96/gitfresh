package main

import (
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/apolo96/gitfresh"
	"github.com/leaanthony/clir"
)

func main() {
	if err := run(); err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}
}

func run() error {
	/* Logger */
	file, closer, err := gitfresh.NewLogFile(gitfresh.APP_CLI_LOG_FILE)
	if err != nil {
		return err
	}
	defer closer()
	logger := slog.New(slog.NewJSONHandler(file, &slog.HandlerOptions{Level: slog.LevelDebug}))
	logger = logger.With("version", "1.0.0")
	slog.SetDefault(logger)
	/* Config */
	userPath, err := os.UserHomeDir()
	if err != nil {
		slog.Error("error getting user home directory", "error", err.Error())
		return err
	}
	/* Services */
	appOS := &gitfresh.AppOS{}
	gitRepoSvc := gitfresh.NewGitRepositorySvc(logger,
		appOS,
		&gitfresh.FlatFile{
			Name: gitfresh.APP_REPOS_FILE_NAME,
			Path: filepath.Join(userPath, gitfresh.APP_FOLDER),
		},
	)
	agentSvc := gitfresh.NewAgentSvc(logger,
		appOS,
		&gitfresh.FlatFile{
			Name: gitfresh.APP_AGENT_FILE,
			Path: filepath.Join(userPath, gitfresh.APP_FOLDER),
		},
		&http.Client{Timeout: time.Second * 2},
	)
	/* CLI */
	cli := clir.NewCli("gitfresh", "A DX Tool to keep the git repositories updated 😎", "v1.0.0")
	cli.NewSubCommandFunction("config", "Configure the application parameters", configCmd)
	/* Init Command */
	initCommand := cli.NewSubCommand("init", "Initialise the workspace and agent")
	flags := &AppFlags{}
	initCommand.AddFlags(flags)
	initCommand.Action(func() error {
		return initCmd(gitRepoSvc, agentSvc)
	})
	/* Scan Command */
	scan := cli.NewSubCommand("scan", "Discover new repositories to refresh")
	scan.Action(func() error {
		return scanCmd(gitRepoSvc)
	})
	/* Status Command */
	status := cli.NewSubCommand("status", "Check agent status")
	status.Action(func() error {
		return statusCmd(agentSvc)
	})
	return cli.Run()
}
