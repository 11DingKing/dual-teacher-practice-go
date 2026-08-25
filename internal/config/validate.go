package config

import (
	"fmt"
	"strings"
	"time"
)

func (c Config) Validate() error {
	if strings.TrimSpace(c.HTTPAddr) == "" {
		return fmt.Errorf("http address is required")
	}
	if strings.TrimSpace(c.DBPath) == "" {
		return fmt.Errorf("database path is required")
	}
	if c.SessionTTL < time.Minute {
		return fmt.Errorf("session ttl too short")
	}
	if c.WorkerInterval <= 0 {
		return fmt.Errorf("worker interval must be positive")
	}
	return nil
}
func (c Config) IsEphemeral() bool {
	return c.DBPath == ":memory:" || strings.HasPrefix(c.DBPath, "file::memory:")
}
