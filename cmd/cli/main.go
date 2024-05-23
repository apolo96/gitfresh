package main

import (
	"fmt"
	"io"
	"log/slog"
	"os"

	"github.com/leaanthony/clir"
)

var devMode string = "off"

func main() {
	if err := run(os.Stdout, os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "\n%s\n", "error: "+err.Error())
		os.Exit(1)
	}
}

func run(w io.Writer, args []string) error {
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
	/* Version Command */
	cli.NewSubCommandFunction("version", "Show cli and agent version", func(_ *struct{}) error {
		renderText(w, "🌟 gitfresh version 1.0.0 \n   A Developer Experience Tool to keep the git repositories updated")
		return nil
	})
	/* Config Command */
	config := cli.NewSubCommand("config", "Configure the application parameters")
	config.AddFlags(flags)
	config.Action(func() error {
		return configCmd(svcProvider.appConfig, flags)
	})
	/* Init Command */
	initCommand := cli.NewSubCommand("init", "Initialise the Workspace and Agent")
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
	status := cli.NewSubCommand("status", "Check Agent Status")
	status.Action(func() error {
		return statusCmd(svcProvider.agent)
	})
	/* Start Command */
	start := cli.NewSubCommand("start", "Start the Agent")
	start.Action(func() error {
		return startCmd(svcProvider.agent)
	})
	/* Stop Command */
	stop := cli.NewSubCommand("stop", "Stop the Agent")
	stop.Action(func() error {
		return stopCmd(svcProvider.agent)
	})
	return cli.Run(args...)
}
