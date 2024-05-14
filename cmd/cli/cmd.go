package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
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
		flags.TunnelToken = PromptSecret("Type the TunnelToken:", true)
	}
	if flags.GitServerToken == "" {
		flags.GitServerToken = PromptSecret("Type the GitServerToken:", true)
	}
	if flags.TunnelDomain == "" {
		flags.TunnelDomain = PromptSecret("Type the TunnelDomain:", false)
	}
	if flags.GitWorkDir == "" {
		workdir, err := os.Getwd()
		if err != nil {
			slog.Error(err.Error())
			return err
		}
		if !PromptConfirm("Type Y/N to confirm", flags.GitWorkDir) {
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
	file, err := gitfresh.CreateConfigFile((*gitfresh.AppFlags)(flags))
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
	repos, err := gitfresh.ScanRepositories(config.GitWorkDir, "github.com")
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
	/* Start Agent */
	ok, err := gitfresh.IsAgentRunning()
	tick := time.NewTicker(time.Microsecond)
	if !ok && err != nil {
		println("Loading GitFresh Agent...")
		cmd := exec.Command("./api")
		if err := cmd.Start(); err != nil {
			return err
		}
		slog.Info("gitfresh agent process", "id", cmd.Process.Pid)
		gitfresh.SaveAgentPID(cmd.Process.Pid)
		tick.Reset(time.Second * 3)
	}
	/* Status check */
	println("Check GitFresh Agent Status...")
	req, err := http.NewRequest("GET", "http://127.0.0.1:9191", &bytes.Buffer{})
	if err != nil {
		slog.Error(err.Error())
		return err
	}
	client := &http.Client{}
	var respBody io.ReadCloser
	times := 5
	for {
		<-tick.C
		println("Checking agent status ...")
		if times <= 0 {
			return errors.New("timeout checking agent status")
		}
		resp, err := client.Do(req)
		if err != nil {
			slog.Error(err.Error())
			continue
		}
		if resp.StatusCode == http.StatusOK {
			respBody = resp.Body
			tick.Stop()
			break
		}
		times--
		slog.Error(resp.Status)
	}
	var agent struct {
		ApiVersion   string `json:"api_version"`
		TunnelDomain string `json:"tunnel_domain"`
	}
	body, _ := io.ReadAll(respBody)
	if err := json.Unmarshal(body, &agent); err != nil {
		slog.Error(err.Error())
		return err
	}
	println("GitFresh Agent is running")
	if config.TunnelDomain == "" {
		config.TunnelDomain = agent.TunnelDomain
		//gitfresh.CreateConfigFile(config)
	}
	fRepos := []*gitfresh.Repository{}
	for i, r := range repos {
		if err := gitfresh.CreateGitServerHook(&r, config); err != nil {
			slog.Error(err.Error())
			continue
		}
		fRepos = append(fRepos, &repos[i])
	}
	if len(fRepos) < 1 {
		return errors.New("creating webhook for repositories")
	}
	if _, err := gitfresh.SaveReposMetaData(fRepos); err != nil {
		return err
	}
	for _, r := range fRepos {
		url := fmt.Sprintf("https://github.com/apolo96/%s/settings/hooks", r.Name)
		fmt.Printf("Repository: %-30s | URL: %-20s\n", r.Name, url)
	}
	return nil
}

func startcmd(flags *struct{ Verbose bool }) error {
	println("Loading GitFresh Agent...")
	api, _ := os.Getwd()
	slog.Info(api)
	ls, _ := exec.Command("ls", "-l").CombinedOutput()
	slog.Info(string(ls))
	cmd := exec.Command("./api")
	if err := cmd.Start(); err != nil {
		return err
	}
	slog.Info("gitfresh agent process", "id", cmd.Process.Pid)
	gitfresh.SaveAgentPID(cmd.Process.Pid)
	return nil
}
