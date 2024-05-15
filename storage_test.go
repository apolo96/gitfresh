package gitfresh

import (
	"os"
	"path/filepath"
	"testing"

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

func Test_scanRepositories(t *testing.T) {
	type args struct {
		workdir     string
		gitProvider string
	}
	tests := []struct {
		name    string
		args    args
		want    []*GitRepository
		wantErr bool
	}{
		{
			name: "scan repositories successfully",
			args: args{
				workdir:     "/Users/laniakea/code/temp",
				gitProvider: "github.com",
			},
			want: []*GitRepository{
				{
					Owner: "apolo96",
					Name:  "torcli",
				},
				{
					Owner: "apolo96",
					Name:  "metaudio",
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ScanRepositories(tt.args.workdir, tt.args.gitProvider)
			if (err != nil) != tt.wantErr {
				t.Errorf("scanRepositories() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if diff := cmp.Diff(got, tt.want); diff != "" {
				t.Errorf(diff)
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
