package gitfresh

import (
	"fmt"
	"log/slog"
	"math/rand"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
)

var TEST_HOME_DIR, _ = os.UserHomeDir()

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
					GitWorkDir:     filepath.Join(TEST_HOME_DIR, "code"),
					GitHookSecret:  WebHookSecret(),
				},
			},
			wantFile: filepath.Join(TEST_HOME_DIR, APP_FOLDER, APP_CONFIG_FILE_NAME),
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

type MockAppOS struct {
	RunFunc      func(path string, workdir string, args ...string) ([]byte, error)
	LookFunc     func(cmd string) (string, error)
	WalkFuncMock func(path string, fn func(string)) error
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

func TestGitRepositorySvc_ScanRepositories(t *testing.T) {
	comparable := []*GitRepository{}
	rand.New(rand.NewSource(time.Now().UnixNano()))
	num := rand.Intn(10)
	for i := num; i > 0; i-- {
		comparable = append(comparable, &GitRepository{
			Name:  "gitfresh",
			Owner: "apolo96",
		})
	}
	mockAppOS := &MockAppOS{
		RunFunc: func(path string, workdir string, args ...string) ([]byte, error) {
			fmt.Println(path, workdir, args)
			return []byte("https://github.com/apolo96/gitfresh.git"), nil
		},
		LookFunc: func(cmd string) (string, error) {
			return "/bin/cmd", nil
		},
		WalkFuncMock: func(path string, fn func(string)) error {
			for i := range num {
				fn(fmt.Sprint("folder", i))
			}
			return nil
		},
	}
	type fields struct {
		logs  AppLogger
		appOs OSDirCommand
	}
	type args struct {
		workdir     string
		gitProvider string
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
				logs:  slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelDebug, AddSource: true})),
				appOs: mockAppOS,
			},
			args: args{
				"mipc/user/work/code",
				"github.com",
			},
			want:    comparable,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gr := NewGitRepositorySvc(tt.fields.logs, tt.fields.appOs)
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

func Test_saveReposMetaData(t *testing.T) {
	type args struct {
		repos []*GitRepository
	}
	reposFilePath := filepath.Join(TEST_HOME_DIR, APP_FOLDER, APP_REPOS_FILE_NAME)
	tests := []struct {
		name     string
		args     args
		wantFile string
		wantErr  bool
	}{
		{
			name: "save repos sucessfully",
			args: args{
				[]*GitRepository{
					{
						Owner: "apolo96",
						Name:  "torcli",
					},
					{
						Owner: "apolo96",
						Name:  "metaudio",
					},
				},
			},
			wantFile: reposFilePath,
			wantErr:  false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotFile, err := SaveRepositories(tt.args.repos)
			if (err != nil) != tt.wantErr {
				t.Errorf("saveReposMetaData() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotFile != tt.wantFile {
				t.Errorf("saveReposMetaData() = %v, want %v", gotFile, tt.wantFile)
			}
		})
	}
	os.Remove(reposFilePath)
}
