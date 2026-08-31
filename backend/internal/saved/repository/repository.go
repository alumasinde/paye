package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type Repository struct{ DB *sql.DB }

type Snapshot struct {
	PublicID        string          `json:"id"`
	Label           *string         `json:"label"`
	CalculationDate string          `json:"calculation_date"`
	GrossSalary     string          `json:"gross_salary"`
	NetSalary       string          `json:"net_salary"`
	Payload         json.RawMessage `json:"payload"`
	CreatedAt       time.Time       `json:"created_at"`
}

func (r Repository) Save(ctx context.Context, userPublicID string, label *string, date, gross, taxable, pbr, relief, paye, total, net string, stat, custom, rules, payload json.RawMessage) (string, error) {
	id := uuid.NewString()
	_, err := r.DB.ExecContext(ctx,
		`INSERT INTO saved_calculations(public_id,user_id,label,calculation_date,gross_salary,taxable_income,paye_before_relief,personal_relief,paye,statutory_deductions_json,custom_deductions_json,total_deductions,net_salary,rule_versions_json,calculation_snapshot_json)
		 SELECT ?,id,?,?,?,?,?,?,?,?,?,?,?,?,? FROM users WHERE public_id=?`,
		id, label, date, gross, taxable, pbr, relief, paye, stat, custom, total, net, rules, payload, userPublicID)
	return id, err
}

func (r Repository) List(ctx context.Context, user string, limit int) ([]Snapshot, error) {
	rows, err := r.DB.QueryContext(ctx,
		`SELECT sc.public_id, sc.label, DATE_FORMAT(sc.calculation_date,'%Y-%m-%d'), CAST(sc.gross_salary AS CHAR), CAST(sc.net_salary AS CHAR), sc.calculation_snapshot_json, sc.created_at
		 FROM saved_calculations sc JOIN users u ON u.id=sc.user_id
		 WHERE u.public_id=? ORDER BY sc.created_at DESC LIMIT ?`,
		user, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []Snapshot{}
	for rows.Next() {
		var x Snapshot
		if err := rows.Scan(&x.PublicID, &x.Label, &x.CalculationDate, &x.GrossSalary, &x.NetSalary, &x.Payload, &x.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, x)
	}
	return out, rows.Err()
}

func (r Repository) Delete(ctx context.Context, user, id string) error {
	_, err := r.DB.ExecContext(ctx,
		`DELETE sc FROM saved_calculations sc JOIN users u ON u.id=sc.user_id WHERE sc.public_id=? AND u.public_id=?`,
		id, user)
	return err
}

// Rename updates only the label on a saved calculation the user owns. Scoped
// by public_id + the owning user's public_id in the same way Delete is, so
// one person can never rename another person's saved calculation.
func (r Repository) Rename(ctx context.Context, user, id string, label *string) error {
	result, err := r.DB.ExecContext(ctx,
		`UPDATE saved_calculations sc JOIN users u ON u.id=sc.user_id SET sc.label=? WHERE sc.public_id=? AND u.public_id=?`,
		label, id, user)
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return sql.ErrNoRows
	}
	return nil
}
