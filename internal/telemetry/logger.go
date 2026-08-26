package telemetry

import (
	"encoding/json"
	"log"
	"os"
	"time"
)

type Logger struct{ std *log.Logger }

func New() *Logger { return &Logger{std: log.New(os.Stdout, "", 0)} }
func (l *Logger) Event(action string, fields map[string]any) {
	fields["action"] = action
	fields["at"] = time.Now().UTC().Format(time.RFC3339Nano)
	if b, err := json.Marshal(fields); err == nil {
		l.std.Print(string(b))
	}
}
