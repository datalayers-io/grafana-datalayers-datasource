package arrow_flightsql

import (
	"fmt"
	"strings"
)

// Config struct to hold datasource configuration
type config struct {
	Addr     string              `json:"host"`
	Metadata []map[string]string `json:"metadata"`
	Secure   bool                `json:"secure"`
	Username string              `json:"username"`
	Password string              `json:"password"`
	Token    string              `json:"token"`
}

// Validate the configuration
func (cfg config) validate() error {
	if strings.Count(cfg.Addr, ":") == 0 {
		return fmt.Errorf(`server address must be in the form "host:port"`)
	}

	noToken := len(cfg.Token) == 0
	noUserPass := len(cfg.Username) == 0 || len(cfg.Password) == 0

	if noToken && noUserPass && cfg.Secure {
		return fmt.Errorf("token or username/password are required")
	}

	return nil
}
