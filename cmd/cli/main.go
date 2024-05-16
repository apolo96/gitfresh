package main

import (
	"log/slog"
	"os"
	"path/filepath"

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
	/* App Services */
	userPath, err := os.UserHomeDir()
	if err != nil {
		slog.Error("error getting user home directory", "error", err.Error())
		return err
	}
	gitRepoSvc := gitfresh.NewGitRepositorySvc(logger,
		&gitfresh.AppOS{},
		&gitfresh.FlatFile{
			Name: gitfresh.APP_REPOS_FILE_NAME,
			Path: filepath.Join(userPath, gitfresh.APP_FOLDER),
		},
	)
	/* CLI */
	cli := clir.NewCli("gitfresh", "A DX Tool to keep the git repositories updated 😎", "v1.0.0")
	cli.NewSubCommandFunction("config", "Configure the application parameters", configCmd)
	/* Init Command */
	initCommand := cli.NewSubCommand("init", "Initialise the workspace and agent")
	flags := &AppFlags{}
	initCommand.AddFlags(flags)
	initCommand.Action(func() error {
		return initCmd(gitRepoSvc)
	})
	scan := cli.NewSubCommand("scan", "Discover new repositories to refresh")
	scan.Action(func() error {
		return scanCmd(gitRepoSvc)
	})
	cli.NewSubCommandFunction("status", "Check agent status", statusCmd)
	return cli.Run()
}
