package flows

import (
	"raid/infra/internal/config"
	"raid/infra/internal/scaffold"
)

// InitialiseProject is the entry point for `infra init repo` — it loads
// build-time URLs from config and delegates to scaffold.Run.
func InitialiseProject() error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	return scaffold.Run(cfg.GitlabHTTPSDomain, cfg.GitlabSSHDomain)
}
