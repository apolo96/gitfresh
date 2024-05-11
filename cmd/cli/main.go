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
	cli.NewSubCommandFunction("config", "Configure the user integration parameters \n", configCmd)
	cli.NewSubCommandFunction("init", "Initialise the local agent", initCmd)
	cli.NewSubCommandFunction("start", "start", startCmd)
	return cli.Run()
}
