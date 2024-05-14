package main

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/apolo96/gitfresh"
)

type AppFlags struct {
	TunnelToken    string `name:"TunnelToken" description:"Actually gitfresh support only Ngrok Internet Tunnel.\nYou can get a Token going to https://dashboard.ngrok.com/get-started/your-authtoken \n"`
	TunnelDomain   string `name:"TunnelDomain" description:"Actually gitfresh support only Ngrok Internet Tunnel.\nYou can get a Custom Domain going to https://dashboard.ngrok.com/cloud-edge/domains \n"`
	GitServerToken string `name:"GitServerToken" description:"Actually gitfresh support only github.com.\nYou can get a Toke going to https://github.com \n"`
	GitWorkDir     string `name:"GitWorkDir" description:"Your Git working directory where you have all repositories. For example: /users/lio/code . Type the absolute path.\nIf you don't enter a GitWorkDir, then GitFresh assumes that your GitWorkDir is your current directory. \n"`
}

func configCmd(flags *AppFlags) error {
	if flags.TunnelToken == "" {
		flags.TunnelToken = PromptSecret("Type the TunnelToken (Ngrok):", true)
	}
	if flags.GitServerToken == "" {
		flags.GitServerToken = PromptSecret("Type the GitServerToken (Github):", true)
	}
	if flags.TunnelDomain == "" {
		flags.TunnelDomain = PromptSecret("Type the TunnelDomain (Ngrok):", false)
	}
	if flags.GitWorkDir == "" {
		workdir, err := os.Getwd()
		if err != nil {
			slog.Error(err.Error())
			return err
		}
		if !PromptConfirm("Type Y/N to confirm", workdir) {
			workdir = PromptSecret("Type the GitWorkDir:", true)
		}
		flags.GitWorkDir = workdir
	}
	p, err := exec.LookPath("git")
	if err != nil {
		slog.Error("which git path", "error", err.Error())
		println("error:", err.Error())
		println("tip: check that git is installed")
		return err
	}
	if checkGit := strings.ReplaceAll(string(p), "\n", ""); checkGit == "" {
		return errors.New("git is not installed, please install git https://git-scm.com/downloads")
	}
	slog.Info("flags values", "content", fmt.Sprint(flags))
	config := &gitfresh.AppConfig{
		TunnelToken:    flags.TunnelToken,
		TunnelDomain:   flags.TunnelDomain,
		GitServerToken: flags.GitServerToken,
		GitWorkDir:     flags.GitWorkDir,
		GitHookSecret:  gitfresh.WebHookSecret(),
	}
	file, err := gitfresh.CreateConfigFile(config)
	if err != nil {
		slog.Error("creating config file")
		slog.Error(err.Error())
		println("ERROR Creating config file")
		return err
	}
	/* TODO:
	*	Create service for read config File
	 */
	output, err := exec.Command("cat", file).Output()
	if err != nil {
		return err
	}
	os.Stdout.Write(output)
	println("\n\n✅ Config successfully created! Now, run the following command: \n\n gitfresh init \n")
	return nil
}

func initCmd(flags *struct{ Verbose bool }) error {
	config, err := gitfresh.ReadConfigFile()
	if err != nil {
		return err
	}
	repos, err := gitfresh.ScanRepositories(config.GitWorkDir, gitfresh.APP_GIT_PROVIDER)
	if err != nil {
		slog.Error(err.Error())
		return err
	}

	println("Discovery Repositories")
	renderRepos(repos)
	if len(repos) < 1 {
		println("The scanner didn't find available repositories")
		return nil
	}
	/* Start Agent */
	ok, err := gitfresh.IsAgentRunning()
	tick := time.NewTicker(time.Microsecond)
	if !ok && err != nil {
		pid, err := gitfresh.StartAgent()
		if err != nil {
			return err
		}
		gitfresh.SaveAgentPID(pid)
		tick.Reset(time.Second * 3)
	}
	/* Status check */
	println("Check GitFresh Agent Status...")
	agent, err := gitfresh.CheckAgentStatus(tick)
	if err != nil {
		return err
	}
	println("GitFresh Agent is running!")
	if config.TunnelDomain == "" {
		println("Saving TunnelDomain")
		config.TunnelDomain = agent.TunnelDomain
		_, err := gitfresh.CreateConfigFile(config)
		if err != nil {
			return err
		}
	}
	fRepos := []*gitfresh.Repository{}
	for i, r := range repos {
		if err := gitfresh.CreateGitServerHook(r, config); err != nil {
			slog.Error(err.Error())
			continue
		}
		fRepos = append(fRepos, repos[i])
	}
	if len(fRepos) < 1 {
		return errors.New("creating webhook for repositories")
	}
	if _, err := gitfresh.SaveReposMetaData(fRepos); err != nil {
		return err
	}
	println("Tracking Repositories")
	renderRepos(fRepos)
	return nil
}

func refreshCmd(flags *struct{}) error {
	config, err := gitfresh.ReadConfigFile()
	if err != nil {
		return err
	}
	repos, err := gitfresh.ScanRepositories(config.GitWorkDir, gitfresh.APP_GIT_PROVIDER)
	if err != nil {
		slog.Error(err.Error())
		return err
	}
	fRepos := []*gitfresh.Repository{}
	for i, r := range repos {
		if err := gitfresh.CreateGitServerHook(r, config); err != nil {
			slog.Error(err.Error())
			continue
		}
		fRepos = append(fRepos, repos[i])
	}
	if len(fRepos) < 1 {
		return errors.New("creating webhook for repositories")
	}
	if _, err := gitfresh.SaveReposMetaData(fRepos); err != nil {
		return err
	}
	println("New Repositories")
	renderRepos(fRepos)
	return nil
}

func statusCmd(flags *struct{}) error {
	ok, err := gitfresh.IsAgentRunning()
	tick := time.NewTicker(time.Microsecond)
	if !ok {
		println("GitFresh Agent is not running!")
		if err != nil {
			return err
		}
	}
	println("Check GitFresh Agent Status...")
	_, err = gitfresh.CheckAgentStatus(tick)
	if err != nil {
		return err
	}
	println("GitFresh Agent is running!")
	return nil
}
