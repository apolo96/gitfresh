package gitfresh

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"math/rand"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"
)

/* GitServer */
func CreateGitServerHook(repo *GitRepository, config *AppConfig) error {
	url := "https://api.github.com/repos/" + filepath.Join(repo.Owner, repo.Name, "hooks")
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
		slog.Error(err.Error())
		return err
	}
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		slog.Error(err.Error())
		return err
	}
	req.Header.Set("Authorization", "Bearer "+config.GitServerToken)
	req.Header.Set("Accept", "*/*")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")

	client := &http.Client{Timeout: time.Second * 20}
	resp, err := client.Do(req)
	if err != nil {
		slog.Error(err.Error())
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		slog.Info(string(jsonData))
		slog.Info(url)
		rb, _ := io.ReadAll(resp.Body)
		slog.Info(string(rb))
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
				slog.Error(err.Error())
			}
			for _, e := range errResponse.Errors {
				if e.Resource == "Hook" {
					if strings.Contains(e.Message, "already exists") {
						slog.Info(e.Message, "repo", repo.Name)
						return nil
					}
				}
			}
		}
		return errors.New("creating webhook via http, response with " + resp.Status)
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

func DiffRepositories() ([]*GitRepository, error) {
	/* Storage */
	r, _ := ScanRepositories("", "")
	var data []map[string]any
	repos := []*GitRepository{}
	b, err := ListRepository()
	if err != nil {
		return repos, err
	}
	fmt.Println(string(b))
	if err := json.Unmarshal(b, &data); err != nil {
		slog.Error(err.Error())
		return repos, err
	}
	result := make(map[string]struct{})
	for _, v := range data {
		result[v["Name"].(string)] = struct{}{}
	}

	fmt.Println(data)
	for _, rs := range r {
		_, ok := result[rs.Name]
		if ok {
			repos = append(repos, rs)
		}
	}
	return repos, err
}

/* Agent */
func IsAgentRunning() (bool, error) {
	dir, _ := os.UserHomeDir()
	path := filepath.Join(dir, APP_FOLDER, APP_AGENT_FILE)
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false, err
	}
	content, err := os.ReadFile(path)
	if err != nil {
		fmt.Printf("Error al leer el archivo PID: %v\n", err)
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
	process, err := os.FindProcess(pid)
	if err != nil {
		return false, err
	}
	err = process.Signal(os.Signal(syscall.Signal(0)))
	if err != nil {
		return false, err
	}
	return true, nil
}

func SaveAgentPID(pid int) error {
	pidStr := fmt.Sprint(pid)
	dir, _ := os.UserHomeDir()
	path := filepath.Join(dir, APP_FOLDER, APP_AGENT_FILE)
	return os.WriteFile(path, []byte(pidStr), 0644)
}

func StartAgent() (int, error) {
	/* path := exec.LookPath("gitfreshd")
	cmd := exec.Command(path) */
	cmd := exec.Command("./api")
	if err := cmd.Start(); err != nil {
		return 0, err
	}
	slog.Info("gitfresh agent process", "id", cmd.Process.Pid)
	return cmd.Process.Pid, nil
}

type Agent struct {
	ApiVersion   string `json:"api_version"`
	TunnelDomain string `json:"tunnel_domain"`
}

func CheckAgentStatus(tick *time.Ticker) (Agent, error) {
	var agent Agent = Agent{}
	req, err := http.NewRequest("GET", "http://"+API_AGENT_HOST, &bytes.Buffer{})
	if err != nil {
		slog.Error(err.Error())
		return agent, err
	}
	client := &http.Client{}
	var respBody io.ReadCloser
	times := 5
	for {
		<-tick.C
		if times <= 0 {
			return agent, errors.New("timeout checking agent status")
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

	body, _ := io.ReadAll(respBody)
	if err := json.Unmarshal(body, &agent); err != nil {
		slog.Error(err.Error())
		return agent, err
	}
	return agent, nil
}

/* AppConfig */
func CreateConfigFile(config *AppConfig) (file string, err error) {
	dirname, err := os.UserHomeDir()
	if err != nil {
		println("error getting user home directory")
		slog.Error(err.Error())
		return file, err
	}
	content, err := json.MarshalIndent(config, "", "  ")
	slog.Debug("parsing config parameters", "data", string(content))
	if err != nil {
		println("error parsing the config parameters")
		slog.Error(err.Error())
		return file, err
	}
	fl := &FlatFile{Name: APP_CONFIG_FILE_NAME, Path: filepath.Join(dirname, APP_FOLDER)}
	_, err = fl.Write(content)
	if err != nil {
		return filepath.Join(fl.Path, fl.Name), err
	}
	slog.Info("config file created successfully")
	return file, nil
}

func ReadConfigFile() (*AppConfig, error) {
	dirname, err := os.UserHomeDir()
	config := &AppConfig{}
	if err != nil {
		return config, err
	}
	fl := &FlatFile{Name: APP_CONFIG_FILE_NAME, Path: filepath.Join(dirname, APP_FOLDER)}
	file, err := fl.Read()
	if err != nil {
		return config, err
	}
	if err := json.Unmarshal(file, &config); err != nil {
		return config, err
	}
	return config, nil
}

/* GitRepository */
func ScanRepositories(workdir string, gitProvider string) ([]*GitRepository, error) {
	repos := []*GitRepository{}
	dirs, err := os.ReadDir(workdir)
	if err != nil {
		slog.Error(err.Error())
		return repos, err
	}
	for _, f := range dirs {
		if f.IsDir() {
			path := filepath.Join(workdir, f.Name())
			git, _ := exec.LookPath("git")
			c := exec.Command(git, "remote", "get-url", "origin")
			c.Dir = path
			url, err := c.CombinedOutput()
			if err != nil {
				slog.Info(path)
				slog.Error("executing git command", "error", err.Error(), "path", git)
				continue
			}
			slog.Info("repository remote url " + string(url))
			surl := strings.Split(string(url), "/")
			if len(surl) < 4 {
				err = errors.New("getting repository info")
				slog.Error(err.Error())
				continue
			}
			if gitProvider != surl[2] {
				err = errors.New("privider not suported" + surl[2])
				slog.Error(err.Error())
				continue
			}
			repos = append(repos, &GitRepository{Owner: surl[3], Name: strings.ReplaceAll(surl[4], ".git\n", "")})
		}
	}
	return repos, nil
}

func SaveRepositories(repos []*GitRepository) (file string, err error) {
	dirname, err := os.UserHomeDir()
	if err != nil {
		println("error getting user home directory")
		slog.Error(err.Error())
		return file, err
	}
	content, err := json.MarshalIndent(repos, "", "  ")
	slog.Debug("parsing config parameters", "data", string(content))
	if err != nil {
		println("error parsing the config parameters")
		slog.Error(err.Error())
		return file, err
	}
	fl := &FlatFile{Name: APP_REPOS_FILE_NAME, Path: filepath.Join(dirname, APP_FOLDER)}
	_, err = fl.Write(content)
	if err != nil {
		return filepath.Join(fl.Path, fl.Name), err
	}
	return file, nil
}
