package gitfresh

type AppConfig struct {
	TunnelToken    string
	TunnelDomain   string
	GitServerToken string
	GitWorkDir     string
	GitHookSecret  string
}

type GitRepository struct {
	Owner string
	Name  string
}

type Webhook struct {
	Name   string            `json:"name"`
	Active bool              `json:"active"`
	Events []string          `json:"events"`
	Config map[string]string `json:"config"`
}
