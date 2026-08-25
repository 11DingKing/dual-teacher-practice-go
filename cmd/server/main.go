package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"github.com/11DingKing/dual-teacher-practice-go/internal/clock"
	"github.com/11DingKing/dual-teacher-practice-go/internal/config"
	"github.com/11DingKing/dual-teacher-practice-go/internal/domain"
	"github.com/11DingKing/dual-teacher-practice-go/internal/httpapi"
	"github.com/11DingKing/dual-teacher-practice-go/internal/repository"
	"github.com/11DingKing/dual-teacher-practice-go/internal/service"
	"github.com/11DingKing/dual-teacher-practice-go/internal/storage/sqlite"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	cfg := config.Load()
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()
	db, e := sqlite.Open(ctx, cfg.DBPath)
	if e != nil {
		panic(e)
	}
	defer db.Close()
	now := clock.Real{}
	users := repository.Users{DB: db.SQL}
	_ = users.Seed(ctx, domainUser("u-teacher", "teacher", "教师"), hash("teacher-pass"))
	_ = users.Seed(ctx, domainUser("u-hr", "hr", "人事"), hash("hr-pass"))
	auth := service.Auth{Users: users, Sessions: repository.Sessions{DB: db.SQL}, Clock: now, TTL: cfg.SessionTTL}
	srv := httpapi.Server{Auth: auth, Apps: service.Application{DB: db.SQL, Apps: repository.Applications{DB: db.SQL}, Quotas: repository.Quotas{DB: db.SQL}, Audit: repository.Audit{DB: db.SQL}, Clock: now}, Evidence: service.Evidence{Evidence: repository.Evidences{DB: db.SQL}, Apps: repository.Applications{DB: db.SQL}, Audit: repository.Audit{DB: db.SQL}, Clock: now}, Practice: service.Practice{Practices: repository.Practices{DB: db.SQL}, Apps: repository.Applications{DB: db.SQL}, Audit: repository.Audit{DB: db.SQL}, Clock: now}, Review: service.Review{Reviews: repository.Reviews{DB: db.SQL}, Evidences: repository.Evidences{DB: db.SQL}, Apps: repository.Applications{DB: db.SQL}, Audit: repository.Audit{DB: db.SQL}, Clock: now}, Report: service.Report{Apps: repository.Applications{DB: db.SQL}, Audit: repository.Audit{DB: db.SQL}}}
	server := &http.Server{Addr: cfg.HTTPAddr, Handler: srv.Routes(), ReadHeaderTimeout: 5 * time.Second}
	go func() {
		<-ctx.Done()
		shutdownCtx, cancelShutdown := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancelShutdown()
		server.Shutdown(shutdownCtx)
	}()
	fmt.Println("dual-teacher-practice listening on", cfg.HTTPAddr)
	if e := server.ListenAndServe(); e != nil && e != http.ErrServerClosed {
		fmt.Fprintln(os.Stderr, e)
		os.Exit(1)
	}
}
func domainUser(id, role, name string) domain.User {
	return domain.User{ID: id, Username: role, Role: domain.Role(role), DisplayName: name, CreatedAt: time.Now().UTC()}
}
func hash(v string) string { h := sha256.Sum256([]byte(v)); return hex.EncodeToString(h[:]) }
