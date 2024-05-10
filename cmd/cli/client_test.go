package main

import "testing"

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
			name: "creating webhook failed",
			args: args{
				&Repository{
					Owner: "apolo96",
					Name:  "meataudio",
				},
				&AppConfig{
					TunnelToken:    "12312322",
					TunnelDomain:   "wwww.apolo906.com",
					GitServerToken: "090909090",
					GitWorkDir:     "",
					GitHookSecret:  "4939292",
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := createGitServerHook(tt.args.repo, tt.args.config); (err != nil) != tt.wantErr {
				t.Errorf("createGitServerHook() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
