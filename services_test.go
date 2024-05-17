package gitfresh

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
)

/* MockFlatFile */
type MockFlatFile struct {
	Name      string
	Path      string
	WriteFunc func(data []byte) (n int, err error)
	ReadFunc  func() (n []byte, err error)
}

func (f *MockFlatFile) Write(data []byte) (n int, err error) {
	return f.WriteFunc(data)
}

func (f *MockFlatFile) Read() (n []byte, err error) {
	return f.ReadFunc()
}

/* MockAppOS */
type MockAppOS struct {
	RunFunc          func(path string, workdir string, args ...string) ([]byte, error)
	LookFunc         func(cmd string) (string, error)
	WalkFuncMock     func(path string, fn func(string)) error
	StartFunc        func(path string, workdir string, args ...string) (int, error)
	UserHomePathFunc func() (string, error)
	FindProgramFunc  func(pid int) (bool, error)
}

func (m *MockAppOS) StartProgram(path string, workdir string, args ...string) (int, error) {
	return m.StartFunc(path, workdir, args...)
}

func (m *MockAppOS) UserHomePath() (string, error) {
	return m.UserHomePathFunc()
}

func (m *MockAppOS) FindProgram(pid int) (bool, error) {
	return m.FindProgramFunc(pid)
}

func (m *MockAppOS) RunProgram(path string, workdir string, args ...string) ([]byte, error) {
	return m.RunFunc(path, workdir, args...)
}

func (m *MockAppOS) LookProgram(cmd string) (string, error) {
	return m.LookFunc(cmd)
}

func (m *MockAppOS) WalkDirFunc(path string, fn func(string)) error {
	return m.WalkFuncMock(path, fn)
}

/* MockClient */
type MockClient struct {
}

func (m *MockClient) Do(req *http.Request) (*http.Response, error) {
	return &http.Response{
		StatusCode: 200,
		Body:       io.NopCloser(strings.NewReader(fmt.Sprintf(`{"api_version":"1.0.0", "tunnel_domain":"%s"}`, tunnelURL))),
	}, nil
}

/* Table Tests */
type fields struct {
	logs      AppLogger
	appOs     OSDirCommand
	fileStore FlatFiler
}

var T_USERHOME_DIR, _ = os.UserHomeDir()
var mockAppOS *MockAppOS
var tnum = rand.New(rand.NewSource(time.Now().UnixNano())).Intn(10)
var storePath = filepath.Join(T_USERHOME_DIR, APP_FOLDER)
var tfileStoreRepo = &FlatFile{Name: APP_REPOS_FILE_NAME, Path: storePath}

/* Agent SVC */
var tPID = 57546
var tunnelURL string = "refreh-webhok-tunnerl.com"
var tFileStoreAgent = &MockFlatFile{
	Name: APP_AGENT_FILE,
	Path: storePath,
	WriteFunc: func(data []byte) (n int, err error) {
		return len([]byte(fmt.Sprint(tPID))), nil
	},
	ReadFunc: func() (n []byte, err error) {
		return []byte(fmt.Sprint(tPID)), nil
	},
}

func TestMain(m *testing.M) {
	/* Global Arrange */
	mockAppOS = &MockAppOS{
		RunFunc: func(path string, workdir string, args ...string) ([]byte, error) {
			fmt.Println(path, workdir, args)
			return []byte("https://github.com/apolo96/gitfresh.git"), nil
		},
		LookFunc: func(cmd string) (string, error) {
			return "/bin/cmd", nil
		},
		WalkFuncMock: func(path string, fn func(string)) error {
			for i := range tnum {
				fn(fmt.Sprint("folder", i))
			}
			return nil
		},
		FindProgramFunc: func(pid int) (bool, error) {
			return true, nil
		},
		StartFunc: func(path, workdir string, args ...string) (int, error) {
			return tPID, nil
		},
		UserHomePathFunc: func() (string, error) {
			return "mipc/user", nil
		},
	}
	/* Global Act */
	out := m.Run()
	os.Exit(out)
}

func Test_createConfigFile(t *testing.T) {
	const token = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	type args struct {
		config *AppConfig
	}
	tests := []struct {
		name     string
		args     args
		wantFile string
		wantErr  bool
	}{
		{
			name: "create config file successfully",
			args: args{
				&AppConfig{
					TunnelToken:    token,
					TunnelDomain:   "",
					GitServerToken: token,
					GitWorkDir:     filepath.Join(T_USERHOME_DIR, "code"),
					GitHookSecret:  WebHookSecret(),
				},
			},
			wantFile: filepath.Join(T_USERHOME_DIR, APP_FOLDER, APP_CONFIG_FILE_NAME),
			wantErr:  false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotFile, err := CreateConfigFile(tt.args.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("createConfigFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotFile != tt.wantFile {
				t.Errorf("createConfigFile() = %v, want %v", gotFile, tt.wantFile)
			}
			os.RemoveAll(gotFile)
		})
	}
}

/* Tests GitRepository SVC */
func TestGitRepositorySvc_ScanRepositories(t *testing.T) {

	type args struct {
		workdir     string
		gitProvider string
	}
	tcomparableRepos := make([]*GitRepository, 0, tnum)
	for i := tnum; i > 0; i-- {
		tcomparableRepos = append(tcomparableRepos, &GitRepository{
			Name:  "gitfresh",
			Owner: "apolo96",
		})
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []*GitRepository
		wantErr bool
	}{
		{
			name: "scan repositories successfully",
			fields: fields{
				logs:      slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelDebug, AddSource: true})),
				appOs:     mockAppOS,
				fileStore: tfileStoreRepo,
			},
			args: args{
				"mipc/user/work/code",
				"github.com",
			},
			want:    tcomparableRepos,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gr := NewGitRepositorySvc(tt.fields.logs, tt.fields.appOs, tt.fields.fileStore)
			got, err := gr.ScanRepositories(tt.args.workdir, tt.args.gitProvider)
			if (err != nil) != tt.wantErr {
				t.Errorf("GitRepositorySvc.ScanRepositories() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if diff := cmp.Diff(got, tt.want); diff != "" {
				t.Error("GitRepositorySvc.ScanRepositories() = ", diff)
			}
		})
	}
}

func TestGitRepositorySvc_SaveRepositories(t *testing.T) {
	type args struct {
		repos []*GitRepository
	}
	jsonData := []byte(`[
		{
		  "Owner": "apolo96",
		  "Name": "torcli"
		},
		{
		  "Owner": "apolo96",
		  "Name": "metaudio"
		}
]`)
	repos := []*GitRepository{}
	if err := json.Unmarshal(jsonData, &repos); err != nil {
		slog.Error("parsing repository data", "error", err.Error())
		return
	}
	tests := []struct {
		name     string
		fields   fields
		args     args
		wantFile int
		wantErr  bool
	}{
		{
			name: "save repos successfully",
			fields: fields{
				logs:      slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelDebug, AddSource: true})),
				appOs:     mockAppOS,
				fileStore: tfileStoreRepo,
			},
			args:     args{repos: repos},
			wantFile: len(jsonData),
			wantErr:  false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gr := NewGitRepositorySvc(tt.fields.logs, tt.fields.appOs, tt.fields.fileStore)
			gotFile, err := gr.SaveRepositories(tt.args.repos)
			if (err != nil) != tt.wantErr {
				t.Errorf("GitRepositorySvc.SaveRepositories() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotFile != int(tt.wantFile) {
				t.Errorf("GitRepositorySvc.SaveRepositories() = %v, want %v", gotFile, tt.wantFile)
			}
		})
	}
}

/* Tests Agent SVC */
func TestAgentSvc_CheckAgentStatus(t *testing.T) {
	/* MockFlatFile for Agent */
	type fields struct {
		logs       AppLogger
		appOS      OSCommander
		fileStore  FlatFiler
		httpClient HttpClienter
	}
	type args struct {
		tick *time.Ticker
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    Agent
		wantErr bool
	}{
		{
			name: "get agent status OK",
			fields: fields{
				logs:       slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelDebug, AddSource: true})),
				appOS:      mockAppOS,
				fileStore:  tFileStoreAgent,
				httpClient: &MockClient{},
			},
			args: args{
				tick: time.NewTicker(time.Millisecond),
			},
			want:    Agent{ApiVersion: "1.0.0", TunnelDomain: tunnelURL},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := AgentSvc{
				logs:       tt.fields.logs,
				appOS:      tt.fields.appOS,
				fileStore:  tt.fields.fileStore,
				httpClient: tt.fields.httpClient,
			}
			got, err := svc.CheckAgentStatus(tt.args.tick)
			if (err != nil) != tt.wantErr {
				t.Errorf("AgentSvc.CheckAgentStatus() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("AgentSvc.CheckAgentStatus() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAgentSvc_StartAgent(t *testing.T) {
	type fields struct {
		logs       AppLogger
		appOS      OSCommander
		fileStore  FlatFiler
		httpClient HttpClienter
	}
	tests := []struct {
		name    string
		fields  fields
		want    int
		wantErr bool
	}{
		{
			name: "start agent succesfully",
			fields: fields{
				logs:       slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelDebug, AddSource: true})),
				appOS:      mockAppOS,
				fileStore:  tFileStoreAgent,
				httpClient: &MockClient{},
			},
			want: tPID,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := AgentSvc{
				logs:       tt.fields.logs,
				appOS:      tt.fields.appOS,
				fileStore:  tt.fields.fileStore,
				httpClient: tt.fields.httpClient,
			}
			got, err := svc.StartAgent()
			if (err != nil) != tt.wantErr {
				t.Errorf("AgentSvc.StartAgent() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("AgentSvc.StartAgent() = %v, want %v", got, tt.want)
			}
		})
	}
}

/* Tests GitServer SVC */
func Test_createGitServerHook(t *testing.T) {
	type args struct {
		repo   *GitRepository
		config *AppConfig
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "creating webhook success",
			args: args{
				&GitRepository{
					Owner: "apolo96",
					Name:  "docker-php-7.4-nginx-dev",
				},
				&AppConfig{
					TunnelToken:    os.Getenv("NGROK_TOKEN"),
					TunnelDomain:   os.Getenv("NGROK_DOMAIN"),
					GitServerToken: os.Getenv("GITHUB_TOKEN"),
					GitWorkDir:     "",
					GitHookSecret:  "GITFRESH010231",
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := CreateGitServerHook(tt.args.repo, tt.args.config); (err != nil) != tt.wantErr {
				t.Errorf("createGitServerHook() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
