package main

import (
	"log/slog"
	"os"

	"github.com/leaanthony/clir"
)

func main() {
	if err := run(); err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}
}

func run() error {
	cli := clir.NewCli("gitfresh", "A DX Tool to keep the git repositories updated 😎", "v1.0.0")
	cli.NewSubCommandFunction("config", "Configure the application parameters", configCmd)
	cli.NewSubCommandFunction("init", "Initialise the workspace and agent", initCmd)
	cli.NewSubCommandFunction("refresh", "Add new repositories for refreshing", refreshCmd)
	cli.NewSubCommandFunction("status", "Check status agent", statusCmd)
	return cli.Run()
}
