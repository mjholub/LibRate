// group functions concerned with parsing config for non-core infra
package cfg

import "os/exec"

func getPGConfigPath() string {
	pgConfig, err := exec.LookPath("pg_config")
	if err != nil {
		log.Error().Err(err).Msg("Failed to find pg_config. Is Postgres installed?")
		return ""
	}
	return pgConfig
}
