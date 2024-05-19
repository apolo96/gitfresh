package main

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/leaanthony/clir"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}

func run() error {
	svcProvider, err := NewServiceProvider()
	slog.SetDefault(svcProvider.logger.log)
	defer svcProvider.logger.closer()
	if err != nil {
		println("error: loading service provider")
		return err
	}
	/* CLI */
	cli := clir.NewCli("gitfresh", "A DX Tool to keep the git repositories updated 😎", "v1.0.0")
	flags := &AppFlags{}
	/* Config Command */
	config := cli.NewSubCommand("config", "Configure the application parameters")
	config.AddFlags(flags)
	config.Action(func() error {
		return configCmd(svcProvider.appConfig, flags)
	})
	/* Init Command */
	initCommand := cli.NewSubCommand("init", "Initialise the workspace and agent")
	initCommand.Action(func() error {
		return initCmd(
			svcProvider.gitRepository,
			svcProvider.agent,
			svcProvider.appConfig,
			svcProvider.gitServer,
		)
	})
	/* Scan Command */
	scan := cli.NewSubCommand("scan", "Discover new repositories to refresh")
	scan.Action(func() error {
		return scanCmd(
			svcProvider.gitRepository,
			svcProvider.appConfig,
			svcProvider.gitServer,
		)
	})
	/* Status Command */
	status := cli.NewSubCommand("status", "Check agent status")
	status.Action(func() error {
		return statusCmd(svcProvider.agent)
	})
	return cli.Run()
}
