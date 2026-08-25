package repository

import (
	"context"
	"database/sql"
	"github.com/11DingKing/dual-teacher-practice-go/internal/domain"
	"time"
)

type Audit struct{ DB *sql.DB }

func (r Audit) Append(ctx context.Context, e domain.AuditEvent) error {
	_, x := r.DB.ExecContext(ctx, "INSERT INTO audit_events(id,actor_id,action,entity_type,entity_id,outcome,request_id,details,created_at) VALUES(?,?,?,?,?,?,?,?,?)", e.ID, e.ActorID, e.Action, e.EntityType, e.EntityID, e.Outcome, e.RequestID, e.Details, e.CreatedAt.Format(time.RFC3339Nano))
	return x
}
func (r Audit) ForEntity(ctx context.Context, typ, id string) ([]domain.AuditEvent, error) {
	rows, e := r.DB.QueryContext(ctx, "SELECT id,actor_id,action,entity_type,entity_id,outcome,request_id,details,created_at FROM audit_events WHERE entity_type=? AND entity_id=? ORDER BY created_at", typ, id)
	if e != nil {
		return nil, e
	}
	defer rows.Close()
	var out []domain.AuditEvent
	for rows.Next() {
		var a domain.AuditEvent
		if e := rows.Scan(&a.ID, &a.ActorID, &a.Action, &a.EntityType, &a.EntityID, &a.Outcome, &a.RequestID, &a.Details, &a.CreatedAt); e != nil {
			return nil, e
		}
		out = append(out, a)
	}
	return out, rows.Err()
}
