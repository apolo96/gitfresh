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

type Agent struct {
	ApiVersion   string `json:"api_version"`
	TunnelDomain string `json:"tunnel_domain"`
}

/* API */

type APIRepository struct {
	Name string `json:"name"`
}

type APIPayload struct {
	Ref        string        `json:"ref"`
	Repository APIRepository `json:"repository"`
	Commit     string        `json:"after"`
}
