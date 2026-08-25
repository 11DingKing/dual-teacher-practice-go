package config

import (
	"testing"
	"time"
)

func TestConfigValidation(t *testing.T) {
	base := Config{HTTPAddr: ":8080", DBPath: "db", SessionTTL: time.Hour, WorkerInterval: time.Second}
	if e := base.Validate(); e != nil {
		t.Fatal(e)
	}
	cases := []Config{{DBPath: "db", SessionTTL: time.Hour, WorkerInterval: time.Second}, {HTTPAddr: ":8080", SessionTTL: time.Hour, WorkerInterval: time.Second}, {HTTPAddr: ":8080", DBPath: "db", SessionTTL: time.Second, WorkerInterval: time.Second}, {HTTPAddr: ":8080", DBPath: "db", SessionTTL: time.Hour}}
	for i, c := range cases {
		if e := c.Validate(); e == nil {
			t.Fatalf("case %d accepted", i)
		}
	}
}
func TestEphemeralPaths(t *testing.T) {
	for _, path := range []string{":memory:", "file::memory:?cache=shared"} {
		c := Config{DBPath: path}
		if !c.IsEphemeral() {
			t.Fatal(path)
		}
	}
	if (Config{DBPath: "./app.db"}).IsEphemeral() {
		t.Fatal("disk path marked ephemeral")
	}
}
