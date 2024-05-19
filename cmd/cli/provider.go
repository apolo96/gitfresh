package main

import (
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/apolo96/gitfresh"
)

func serviceProvider() (
	*gitfresh.GitServerSvc,
	*gitfresh.AgentSvc,
	*gitfresh.AppConfigSvc,
	*gitfresh.GitRepositorySvc,
	*slog.Logger,
	error,
) {
	/* Logger */
	file, closer, err := gitfresh.NewLogFile(gitfresh.APP_CLI_LOG_FILE)
	if err != nil {
		return nil, nil, nil, nil, nil, err
	}
	defer closer()
	logger := slog.New(slog.NewJSONHandler(file, &slog.HandlerOptions{Level: slog.LevelDebug}))
	logger = logger.With("version", "1.0.0")
	/* Config */
	userPath, err := os.UserHomeDir()
	if err != nil {
		slog.Error("error getting user home directory", "error", err.Error())
		return nil, nil, nil, nil, nil, err
	}
	path := filepath.Join(userPath, gitfresh.APP_FOLDER)
	/* Services Provider */
	appOS := &gitfresh.AppOS{}
	gitRepoSvc := gitfresh.NewGitRepositorySvc(logger,
		appOS,
		&gitfresh.FlatFile{
			Name: gitfresh.APP_REPOS_FILE_NAME,
			Path: path,
		},
	)
	agentSvc := gitfresh.NewAgentSvc(logger,
		appOS,
		&gitfresh.FlatFile{
			Name: gitfresh.APP_AGENT_FILE,
			Path: path,
		},
		&http.Client{Timeout: time.Second * 2},
	)
	appConfigSvc := gitfresh.NewAppConfigSvc(logger, &gitfresh.FlatFile{Name: gitfresh.APP_CONFIG_FILE_NAME, Path: path})
	gitServerSvc := gitfresh.NewGitServerSvc(logger, &http.Client{Timeout: time.Second * 3})
	return gitServerSvc, agentSvc, appConfigSvc, gitRepoSvc, logger, nil
}
