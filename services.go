package gitfresh

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"math/rand"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

const devModeOff = "Off"

var DevMode string = devModeOff

/* GitServer */
type GitServerSvc struct {
	logs       AppLogger
	httpClient HttpClienter
}

func NewGitServerSvc(l AppLogger, c HttpClienter) *GitServerSvc {
	return &GitServerSvc{
		logs:       l,
		httpClient: c,
	}
}

func (svc GitServerSvc) CreateGitServerHook(repo *GitRepository, config *AppConfig) error {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/hooks", repo.Owner, repo.Name)
	if !strings.Contains(config.TunnelDomain, "https://") {
		config.TunnelDomain = "https://" + config.TunnelDomain
	}
	webhook := Webhook{
		Name:   "web",
		Active: true,
		Events: []string{"push"},
		Config: map[string]string{
			"url":          config.TunnelDomain,
			"content_type": "json",
			"secret":       config.GitHookSecret,
			"insecure_ssl": "0",
		},
	}
	jsonData, err := json.Marshal(webhook)
	if err != nil {
		svc.logs.Error(err.Error())
		return err
	}
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		svc.logs.Error(err.Error())
		return err
	}
	req.Header.Set("Authorization", "Bearer "+config.GitServerToken)
	req.Header.Set("Accept", "*/*")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")

	resp, err := svc.httpClient.Do(req)
	if err != nil {
		svc.logs.Error(err.Error())
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		svc.logs.Info(string(jsonData))
		svc.logs.Info(url)
		rb, _ := io.ReadAll(resp.Body)
		svc.logs.Info(string(rb))
		if resp.StatusCode == http.StatusUnprocessableEntity {
			var errResponse struct {
				Message string `json:"message"`
				Errors  []struct {
					Resource string `json:"resource"`
					Code     string `json:"code"`
					Message  string `json:"message"`
				}
				DocumentationURL string `json:"documentation_url"`
			}
			err := json.Unmarshal(rb, &errResponse)
			if err != nil {
				svc.logs.Error(err.Error())
			}
			for _, e := range errResponse.Errors {
				if e.Resource == "Hook" {
					if strings.Contains(e.Message, "already exists") {
						svc.logs.Info(e.Message, "repo", repo.Name)
						return nil
					}
				}
			}
		}
		return errors.New("creating webhook via http, response with " + resp.Status)
	}
	return nil
}

/* Agent */
type AgentSvc struct {
	logs       AppLogger
	appOS      OSCommander
	fileStore  FlatFiler
	httpClient HttpClienter
}

func NewAgentSvc(l AppLogger, a OSCommander, f FlatFiler, c HttpClienter) *AgentSvc {
	return &AgentSvc{
		logs:       l,
		appOS:      a,
		fileStore:  f,
		httpClient: c,
	}
}

func (svc AgentSvc) IsAgentRunning() (bool, error) {
	content, err := svc.fileStore.Read()
	if err != nil {
		return false, err
	}
	pidstr := strings.TrimSpace(string(content))
	if pidstr == "" {
		return false, err
	}
	pid, err := strconv.Atoi(pidstr)
	if err != nil {
		fmt.Println("Error during conversion")
		return false, err
	}
	return svc.appOS.FindProgram(pid)
}

func (svc AgentSvc) StopAgent() error {
	content, err := svc.fileStore.Read()
	if err != nil {
		return err
	}
	pidstr := strings.TrimSpace(string(content))
	if pidstr == "" {
		return err
	}
	pid, err := strconv.Atoi(pidstr)
	if err != nil {
		fmt.Println("Error during conversion")
		return err
	}
	return svc.appOS.StopProgram(pid)
}

func (svc AgentSvc) SaveAgentPID(pid int) (int, error) {
	return svc.fileStore.Write([]byte(fmt.Sprint(pid)))
}

func (svc AgentSvc) StartAgent() (int, error) {
	slog.Info("Application DevMode " + DevMode)
	var path string = "./api"
	if DevMode == devModeOff {
		p, err := svc.appOS.LookProgram("gitfreshd")
		if err != nil {
			slog.Error("getting agent os path", "error", err.Error())
			return 0, err
		}
		path = p
	}
	pid, err := svc.appOS.StartProgram(path, []string{}...)
	if err != nil {
		slog.Error("starting agent", "error", err.Error())
		return 0, err
	}
	slog.Info("running agent", "pid", pid)
	return pid, nil
}

func (svc AgentSvc) CheckAgentStatus(tick *time.Ticker) (Agent, error) {
	var agent Agent = Agent{}
	req, err := http.NewRequest("GET", "http://"+API_AGENT_HOST, &bytes.Buffer{})
	if err != nil {
		slog.Error(err.Error())
		return agent, err
	}
	var respBody io.ReadCloser
	var resp *http.Response
	times := 3
	for {
		<-tick.C
		if times <= 0 {
			return agent, errors.New("timeout checking agent status")
		}
		resp, err = svc.httpClient.Do(req)
		if err != nil {
			times--
			slog.Error(err.Error())
			continue
		}
		break
	}
	if resp.StatusCode != http.StatusOK {
		slog.Error("error response", "status", resp.Status)
		return agent, errors.New("http response " + resp.Status)
	}
	respBody = resp.Body
	defer resp.Body.Close()
	tick.Stop()
	body, _ := io.ReadAll(respBody)
	if err := json.Unmarshal(body, &agent); err != nil {
		slog.Error(err.Error())
		return agent, err
	}
	return agent, nil
}

/* AppConfig */
type AppConfigSvc struct {
	logs      AppLogger
	fileStore FlatFiler
}

func NewAppConfigSvc(l AppLogger, f FlatFiler) *AppConfigSvc {
	return &AppConfigSvc{
		logs:      l,
		fileStore: f,
	}
}

func (svc AppConfigSvc) CreateConfigFile(config *AppConfig) error {
	content, err := json.MarshalIndent(config, "", "  ")
	svc.logs.Debug("parsing config parameters", "data", string(content))
	if err != nil {
		println("error parsing the config parameters")
		svc.logs.Error(err.Error())
		return err
	}
	_, err = svc.fileStore.Write(content)
	if err != nil {
		return err
	}
	svc.logs.Info("config file created successfully")
	return nil
}

func (svc AppConfigSvc) ReadConfigFile() (*AppConfig, error) {
	config := &AppConfig{}
	file, err := svc.fileStore.Read()
	if err != nil {
		return config, err
	}
	if err := json.Unmarshal(file, &config); err != nil {
		return config, err
	}
	return config, nil
}

/* GitRepository */
type GitRepositorySvc struct {
	logs      AppLogger
	appOS     OSDirCommand
	fileStore FlatFiler
}

func NewGitRepositorySvc(l AppLogger, a OSDirCommand, f FlatFiler) *GitRepositorySvc {
	return &GitRepositorySvc{
		logs:      l,
		appOS:     a,
		fileStore: f,
	}
}

func (gr GitRepositorySvc) ScanRepositories(workdir string, gitProvider string) ([]*GitRepository, error) {
	repos := []*GitRepository{}
	fn := func(dirname string) {
		workdir := filepath.Join(workdir, dirname)
		git, _ := gr.appOS.LookProgram("git")
		url, err := gr.appOS.RunProgram(git, workdir, "remote", "get-url", "origin")
		if err != nil {
			gr.logs.Error("executing git command", "error", err.Error(), "path", git, "workdir", workdir)
			return
		}
		gr.logs.Info("repository remote url " + string(url))
		surl := strings.Split(string(url), "/")
		if len(surl) < 4 {
			err = errors.New("getting repository info")
			gr.logs.Error(err.Error())
			return
		}
		if gitProvider != surl[2] {
			err = errors.New("privider not suported" + surl[2])
			gr.logs.Error(err.Error())
			return
		}
		name := strings.ReplaceAll(strings.ReplaceAll(surl[4], ".git", ""), "\n", "")
		repos = append(repos, &GitRepository{Owner: surl[3], Name: name})
	}
	err := gr.appOS.WalkDirFunc(workdir, fn)
	if err != nil {
		gr.logs.Error(err.Error())
		return repos, err
	}
	return repos, nil
}

func (gr GitRepositorySvc) SaveRepositories(repos []*GitRepository) (n int, err error) {
	content, err := json.MarshalIndent(repos, "", "  ")
	slog.Debug("parsing config parameters", "data", string(content))
	if err != nil {
		println("error parsing the config parameters")
		slog.Error(err.Error())
		return 0, err
	}
	n, err = gr.fileStore.Write(content)
	if err != nil {
		return n, err
	}
	return n, nil
}

func (gr GitRepositorySvc) Pull(workdir, repoName, branch string) error {
	git, err := gr.appOS.LookProgram("git")
	if err != nil {
		slog.Error("which git path", "error", err.Error())
		return err
	}
	workspace := filepath.Join(workdir, repoName)
	out, err := gr.appOS.RunProgram(git, workspace, "pull", "origin", branch)
	if err != nil {
		gr.logs.LogAttrs(
			context.Background(),
			slog.LevelError,
			"executing git command",
			slog.String("error", err.Error()),
			slog.String("path", git),
			slog.String("workspace", workspace),
			slog.String("stdout", string(out)),
		)
		return err
	}
	return nil
}

func WebHookSecret() string {
	const alpha = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	rand.New(rand.NewSource(time.Now().UnixNano()))
	secret := make([]byte, 10)
	for i := range secret {
		secret[i] = alpha[rand.Intn(len(alpha))]
	}
	return string(secret)
}
