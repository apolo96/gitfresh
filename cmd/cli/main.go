package main

import (
	"log/slog"
	"os"

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
	/* CLI */
	cli := clir.NewCli("gitfresh", "A DX Tool to keep the git repositories updated 😎", "v1.0.0")
	cli.NewSubCommandFunction("config", "Configure the application parameters", configCmd)
	cli.NewSubCommandFunction("init", "Initialise the workspace and agent", initCmd)
	cli.NewSubCommandFunction("scan", "Discover new repositories to refresh", scanCmd)
	cli.NewSubCommandFunction("status", "Check agent status", statusCmd)
	return cli.Run()
}
