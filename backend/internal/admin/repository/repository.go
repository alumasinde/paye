package repository

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"time"

	"github.com/alumasinde/budget254-paye-api/internal/admin/model"
)

type Repository struct{ DB *sql.DB }

func (r Repository) ByEmail(ctx context.Context, email string) (model.AdminUser, error) {
	var u model.AdminUser
	err := r.DB.QueryRowContext(ctx, `SELECT id,public_id,email,password_hash,first_name,last_name,status,created_at,failed_login_count,locked_until FROM admin_users WHERE email=?`, email).
		Scan(&u.ID, &u.PublicID, &u.Email, &u.PasswordHash, &u.FirstName, &u.LastName, &u.Status, &u.CreatedAt, &u.FailedLoginCount, &u.LockedUntil)
	if err != nil {
		return u, err
	}
	if err := r.loadRoles(ctx, &u); err != nil {
		return u, err
	}
	return u, nil
}

func (r Repository) ByPublicID(ctx context.Context, publicID string) (model.AdminUser, error) {
	var u model.AdminUser
	err := r.DB.QueryRowContext(ctx, `SELECT id,public_id,email,password_hash,first_name,last_name,status,created_at,failed_login_count,locked_until FROM admin_users WHERE public_id=?`, publicID).
		Scan(&u.ID, &u.PublicID, &u.Email, &u.PasswordHash, &u.FirstName, &u.LastName, &u.Status, &u.CreatedAt, &u.FailedLoginCount, &u.LockedUntil)
	if err != nil {
		return u, err
	}
	if err := r.loadRoles(ctx, &u); err != nil {
		return u, err
	}
	return u, nil
}

func (r Repository) loadRoles(ctx context.Context, u *model.AdminUser) error {
	rows, err := r.DB.QueryContext(ctx, `SELECT DISTINCT ar.code, ap.code
        FROM admin_user_roles aur
        JOIN admin_roles ar ON ar.id=aur.role_id
        LEFT JOIN admin_role_permissions arp ON arp.role_id=ar.id
        LEFT JOIN admin_permissions ap ON ap.id=arp.permission_id
        WHERE aur.admin_user_id=?`, u.ID)
	if err != nil {
		return err
	}
	defer rows.Close()
	seenR, seenP := map[string]bool{}, map[string]bool{}
	for rows.Next() {
		var role string
		var permission sql.NullString
		if err := rows.Scan(&role, &permission); err != nil {
			return err
		}
		if !seenR[role] {
			u.Roles = append(u.Roles, role)
			seenR[role] = true
		}
		if permission.Valid && !seenP[permission.String] {
			u.Permissions = append(u.Permissions, permission.String)
			seenP[permission.String] = true
		}
	}
	return rows.Err()
}

// List returns every admin account with its assigned role codes, for the
// admin-user-management screen.
func (r Repository) List(ctx context.Context) ([]model.AdminUser, error) {
	rows, err := r.DB.QueryContext(ctx, `SELECT id,public_id,email,first_name,last_name,status,created_at FROM admin_users ORDER BY created_at`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []model.AdminUser{}
	for rows.Next() {
		var u model.AdminUser
		if err := rows.Scan(&u.ID, &u.PublicID, &u.Email, &u.FirstName, &u.LastName, &u.Status, &u.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, u)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	for i := range out {
		if err := r.loadRoles(ctx, &out[i]); err != nil {
			return nil, err
		}
	}
	return out, nil
}

// Create inserts a new admin account (status ACTIVE) and assigns it the
// given role code (must already exist in admin_roles, e.g. from migration
// 007's seed).
func (r Repository) Create(ctx context.Context, publicID, email, passwordHash, first, last, roleCode string) error {
	tx, err := r.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	res, err := tx.ExecContext(ctx, `INSERT INTO admin_users(public_id,email,password_hash,first_name,last_name,status) VALUES(?,?,?,?,?,'ACTIVE')`, publicID, email, passwordHash, first, last)
	if err != nil {
		return err
	}
	id, _ := res.LastInsertId()
	if _, err := tx.ExecContext(ctx, `INSERT INTO admin_user_roles(admin_user_id,role_id) SELECT ?,id FROM admin_roles WHERE code=?`, id, roleCode); err != nil {
		return err
	}
	return tx.Commit()
}

// SetStatus enables or disables an admin account (status ACTIVE/DISABLED).
func (r Repository) SetStatus(ctx context.Context, publicID, status string) error {
	_, err := r.DB.ExecContext(ctx, `UPDATE admin_users SET status=? WHERE public_id=?`, status, publicID)
	return err
}

func (r Repository) RegisterFailedLogin(ctx context.Context, adminID uint64, lockUntil *time.Time) error {
	_, err := r.DB.ExecContext(ctx, `UPDATE admin_users SET failed_login_count=failed_login_count+1, locked_until=? WHERE id=?`, lockUntil, adminID)
	return err
}
func (r Repository) ResetFailedLogins(ctx context.Context, adminID uint64) error {
	_, err := r.DB.ExecContext(ctx, `UPDATE admin_users SET failed_login_count=0, locked_until=NULL, last_login_at=UTC_TIMESTAMP() WHERE id=?`, adminID)
	return err
}
func (r Repository) UpdatePassword(ctx context.Context, adminID uint64, hash string) error {
	_, err := r.DB.ExecContext(ctx, `UPDATE admin_users SET password_hash=? WHERE id=?`, hash, adminID)
	return err
}

// StoreRefresh records a hashed refresh token for adminID, mirroring
// internal/auth/repository's customer-side StoreRefresh.
func (r Repository) StoreRefresh(ctx context.Context, adminID uint64, raw string, expiry time.Time) error {
	h := sha256.Sum256([]byte(raw))
	_, err := r.DB.ExecContext(ctx, `INSERT INTO admin_refresh_tokens(admin_user_id,token_hash,expires_at) VALUES(?,?,?)`, adminID, hex.EncodeToString(h[:]), expiry)
	return err
}
func (r Repository) AdminByRefresh(ctx context.Context, raw string) (model.AdminUser, error) {
	h := sha256.Sum256([]byte(raw))
	var u model.AdminUser
	err := r.DB.QueryRowContext(ctx, `SELECT au.id,au.public_id,au.email,au.password_hash,au.first_name,au.last_name,au.status,au.created_at,au.failed_login_count,au.locked_until
		FROM admin_refresh_tokens t JOIN admin_users au ON au.id=t.admin_user_id
		WHERE t.token_hash=? AND t.revoked_at IS NULL AND t.expires_at>UTC_TIMESTAMP()`, hex.EncodeToString(h[:])).
		Scan(&u.ID, &u.PublicID, &u.Email, &u.PasswordHash, &u.FirstName, &u.LastName, &u.Status, &u.CreatedAt, &u.FailedLoginCount, &u.LockedUntil)
	if err != nil {
		return u, err
	}
	if err := r.loadRoles(ctx, &u); err != nil {
		return u, err
	}
	return u, nil
}
func (r Repository) RevokeRefresh(ctx context.Context, adminID uint64, raw string) error {
	h := sha256.Sum256([]byte(raw))
	_, err := r.DB.ExecContext(ctx, `UPDATE admin_refresh_tokens SET revoked_at=UTC_TIMESTAMP() WHERE admin_user_id=? AND token_hash=?`, adminID, hex.EncodeToString(h[:]))
	return err
}
