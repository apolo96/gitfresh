package gitfresh

import (
	"os"
	"testing"
)

func Test_createGitServerHook(t *testing.T) {
	type args struct {
		repo   *Repository
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
				&Repository{
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
