package main

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"strings"
	"time"
)

type AppFlags struct {
	TunnelToken    string `name:"TunnelToken" description:"Actually gitfresh support only Ngrok Internet Tunnel.\nYou can get a Token going to https://dashboard.ngrok.com/get-started/your-authtoken \n"`
	TunnelDomain   string `name:"TunnelDomain" description:"Actually gitfresh support only Ngrok Internet Tunnel.\nYou can get a Custom Domain going to https://dashboard.ngrok.com/cloud-edge/domains \n"`
	GitServerToken string `name:"GitServerToken" description:"Actually gitfresh support only github.com.\nYou can get a Toke going to https://github.com \n"`
	GitWorkDir     string `name:"GitWorkDir" description:"Your Git working directory where you have all repositories. For example: /users/lio/code . Type the absolute path.\nIf you don't enter a GitWorkDir, then GitFresh assumes that your GitWorkDir is your current directory. \n"`
}

func startCmd(flags *struct{}) error {
	ok, err := isAgentRunning()
	if ok && err == nil {
		println("GitFresh Agent is running")
		return nil
	}
	cmd := exec.Command("./api")
	if err := cmd.Start(); err != nil {
		return err
	}
	println("Loading GitFresh Agent...")
	time.Sleep(time.Second * 2)
	// check agent status via tunnel
	slog.Info("gitfresh agent process", "id", cmd.Process.Pid)
	saveAgentPID(cmd.Process.Pid)
	return nil
}

func configCmd(flags *AppFlags) error {
	if flags.TunnelToken == "" && flags.GitServerToken == "" && flags.GitWorkDir == "" {
		flags.TunnelToken = PromptSecret("Type the TunnelToken:", true)
		flags.TunnelDomain = PromptSecret("Type the TunnelDomain:", false)
		flags.GitServerToken = PromptSecret("Type the GitServerToken:", true)
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
	flags.GitWorkDir, err = os.Getwd()
	if err != nil {
		println("error: getting current work directory")
		slog.Error(err.Error())
		return err
	}
	fmt.Println("Analizing", flags.TunnelToken, flags.GitServerToken, flags.TunnelDomain, flags.GitWorkDir)
	file, err := createConfigFile(flags)
	if err != nil {
		slog.Error("creating config file")
		slog.Error(err.Error())
		println("ERROR Creating config file")
		return err
	}
	output, err := exec.Command("cat", file).Output()
	if err != nil {
		return err
	}
	os.Stdout.Write(output)
	println("\n\n✅ Config successfully created! Now copy and run the following command: \n\n gitfresh init \n")
	return nil
}

func initCmd(flags *struct{ Verbose bool }) error {
	config, err := readConfigFile()
	if err != nil {
		return err
	}
	repos, err := scanRepositories(config.GitWorkDir, "github.com")
	if err != nil {
		slog.Error(err.Error())
		return err
	}
	if flags.Verbose {
		println("Repositories Scaned")
		for _, r := range repos {
			fmt.Printf("Owner: %-20s | Name: %-20s\n", r.Owner, r.Name)
		}
	}
	if len(repos) < 1 {
		println("The scanner didn't find available repositories")
		return nil
	}
	/* TODO:
	Start WebHook Listener Server
	Pass WEBHOOK_SECRET & CUSTOM DOMAIN (if apply)
	Get TUNNEL DNS
	*/
	fRepos := []*Repository{}
	for i, r := range repos {
		if err := createGitServerHook(&r, config); err != nil {
			continue
		}
		fRepos = append(fRepos, &repos[i])
	}
	if len(fRepos) < 1 {
		return errors.New("creating webhook for repositories")
	}
	if _, err := saveReposMetaData(fRepos); err != nil {
		return err
	}
	for _, r := range fRepos {
		fmt.Printf("Owner: %-20s | Name: %-20s\n", r.Owner, r.Name)
	}
	return nil
}
