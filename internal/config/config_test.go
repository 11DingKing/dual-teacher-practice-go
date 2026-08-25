package config

import (
	"os"
	"testing"
	"time"
)

func TestLoadDefaults(t *testing.T) {
	for _, k := range []string{"HTTP_ADDR", "DB_PATH", "SESSION_TTL", "WORKER_INTERVAL"} {
		os.Unsetenv(k)
	}
	c := Load()
	if c.HTTPAddr != " :8080" && c.HTTPAddr != ":8080" {
		t.Fatal(c.HTTPAddr)
	}
	if c.SessionTTL != 12*time.Hour {
		t.Fatal(c.SessionTTL)
	}
}
func TestLoadEnvironment(t *testing.T) {
	os.Setenv("HTTP_ADDR", ":9090")
	os.Setenv("DB_PATH", "/tmp/a.db")
	os.Setenv("SESSION_TTL", "2h")
	os.Setenv("WORKER_INTERVAL", "5s")
	defer func() {
		for _, k := range []string{"HTTP_ADDR", "DB_PATH", "SESSION_TTL", "WORKER_INTERVAL"} {
			os.Unsetenv(k)
		}
	}()
	c := Load()
	if c.HTTPAddr != ":9090" || c.DBPath != "/tmp/a.db" || c.SessionTTL != 2*time.Hour || c.WorkerInterval != 5*time.Second {
		t.Fatalf("%+v", c)
	}
}
