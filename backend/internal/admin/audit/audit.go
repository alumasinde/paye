package audit

import (
	"context"
	"database/sql"
	"encoding/json"
	"github.com/google/uuid"
)

type Writer struct{ DB *sql.DB }

func (a Writer) Write(ctx context.Context, adminID *uint64, action, entityType, entityID, requestID, ip string, before, after any) error {
	b, _ := json.Marshal(before)
	c, _ := json.Marshal(after)
	_, err := a.DB.ExecContext(ctx, `INSERT INTO admin_audit_logs(public_id,admin_user_id,action,entity_type,entity_public_id,request_id,ip_address,before_json,after_json) VALUES(?,?,?,?,?,?,?,?,?)`, uuid.NewString(), adminID, action, entityType, entityID, requestID, ip, json.RawMessage(b), json.RawMessage(c))
	return err
}

type Entry struct {
	PublicID       string          `json:"id"`
	AdminEmail     *string         `json:"admin_email,omitempty"`
	Action         string          `json:"action"`
	EntityType     string          `json:"entity_type"`
	EntityPublicID *string         `json:"entity_id,omitempty"`
	RequestID      *string         `json:"request_id,omitempty"`
	IPAddress      *string         `json:"ip_address,omitempty"`
	Before         json.RawMessage `json:"before,omitempty"`
	After          json.RawMessage `json:"after,omitempty"`
	CreatedAt      string          `json:"created_at"`
}

// List returns the most recent audit log entries, newest first, joined
// against admin_users for a human-readable actor email (NULL if the admin
// account was since deleted - admin_audit_logs.admin_user_id is ON DELETE
// SET NULL for exactly this reason).
func (a Writer) List(ctx context.Context, limit int) ([]Entry, error) {
	rows, err := a.DB.QueryContext(ctx, `SELECT l.public_id, au.email, l.action, l.entity_type, l.entity_public_id, l.request_id, l.ip_address, l.before_json, l.after_json, DATE_FORMAT(l.created_at,'%Y-%m-%dT%H:%i:%sZ')
		FROM admin_audit_logs l LEFT JOIN admin_users au ON au.id=l.admin_user_id
		ORDER BY l.created_at DESC LIMIT ?`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []Entry{}
	for rows.Next() {
		var e Entry
		var email, entityID, requestID, ip sql.NullString
		if err := rows.Scan(&e.PublicID, &email, &e.Action, &e.EntityType, &entityID, &requestID, &ip, &e.Before, &e.After, &e.CreatedAt); err != nil {
			return nil, err
		}
		if email.Valid {
			e.AdminEmail = &email.String
		}
		if entityID.Valid {
			e.EntityPublicID = &entityID.String
		}
		if requestID.Valid {
			e.RequestID = &requestID.String
		}
		if ip.Valid {
			e.IPAddress = &ip.String
		}
		out = append(out, e)
	}
	return out, rows.Err()
}
