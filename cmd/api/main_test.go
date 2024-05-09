package main

import "testing"

func Test_gitPullCmd(t *testing.T) {
	type args struct {
		workdir  string
		repoName string
		branch   string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			"update repository",
			args{
				workdir:  "/Users/laniakea/code/",
				repoName: "metaudio",
				branch:   "refs/heads/main",
			},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := gitPullCmd(tt.args.workdir, tt.args.repoName, tt.args.branch); (err != nil) != tt.wantErr {
				t.Errorf("gitPullCmd() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
