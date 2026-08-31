package repository

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"time"

	"github.com/alumasinde/budget254-paye-api/internal/auth/model"
)

type Repository struct{ DB *sql.DB }

func (r Repository) CreateUser(ctx context.Context, u model.User) error {
	_, err := r.DB.ExecContext(ctx, `INSERT INTO users(public_id,email,password_hash,first_name,last_name,status) VALUES(?,?,?,?,?,'ACTIVE')`, u.PublicID, u.Email, u.PasswordHash, u.FirstName, u.LastName)
	return err
}
func (r Repository) ByEmail(ctx context.Context, email string) (model.User, error) {
	var u model.User
	err := r.DB.QueryRowContext(ctx, `SELECT id,public_id,email,password_hash,first_name,last_name,status,created_at,failed_login_count,locked_until FROM users WHERE email=?`, email).
		Scan(&u.ID, &u.PublicID, &u.Email, &u.PasswordHash, &u.FirstName, &u.LastName, &u.Status, &u.CreatedAt, &u.FailedLoginCount, &u.LockedUntil)
	return u, err
}
func (r Repository) ByPublicID(ctx context.Context, publicID string) (model.User, error) {
	var u model.User
	err := r.DB.QueryRowContext(ctx, `SELECT id,public_id,email,password_hash,first_name,last_name,status,created_at,failed_login_count,locked_until FROM users WHERE public_id=?`, publicID).
		Scan(&u.ID, &u.PublicID, &u.Email, &u.PasswordHash, &u.FirstName, &u.LastName, &u.Status, &u.CreatedAt, &u.FailedLoginCount, &u.LockedUntil)
	return u, err
}
func (r Repository) RegisterFailedLogin(ctx context.Context, userID uint64, lockUntil *time.Time) error {
	_, err := r.DB.ExecContext(ctx, `UPDATE users SET failed_login_count=failed_login_count+1, locked_until=? WHERE id=?`, lockUntil, userID)
	return err
}
func (r Repository) ResetFailedLogins(ctx context.Context, userID uint64) error {
	_, err := r.DB.ExecContext(ctx, `UPDATE users SET failed_login_count=0, locked_until=NULL, last_login_at=UTC_TIMESTAMP() WHERE id=?`, userID)
	return err
}
func (r Repository) UpdatePassword(ctx context.Context, userID uint64, hash string) error {
	_, err := r.DB.ExecContext(ctx, `UPDATE users SET password_hash=? WHERE id=?`, hash, userID)
	return err
}
func (r Repository) StoreRefresh(ctx context.Context, userID uint64, raw string, expiry time.Time, ua string) error {
	h := sha256.Sum256([]byte(raw))
	_, err := r.DB.ExecContext(ctx, `INSERT INTO user_refresh_tokens(user_id,token_hash,expires_at,user_agent) VALUES(?,?,?,?)`, userID, hex.EncodeToString(h[:]), expiry, ua)
	return err
}
func (r Repository) UserByRefresh(ctx context.Context, raw string) (model.User, error) {
	h := sha256.Sum256([]byte(raw))
	var u model.User
	err := r.DB.QueryRowContext(ctx, `SELECT u.id,u.public_id,u.email,u.password_hash,u.first_name,u.last_name,u.status,u.created_at,u.failed_login_count,u.locked_until
		FROM user_refresh_tokens t JOIN users u ON u.id=t.user_id
		WHERE t.token_hash=? AND t.revoked_at IS NULL AND t.expires_at>UTC_TIMESTAMP()`, hex.EncodeToString(h[:])).
		Scan(&u.ID, &u.PublicID, &u.Email, &u.PasswordHash, &u.FirstName, &u.LastName, &u.Status, &u.CreatedAt, &u.FailedLoginCount, &u.LockedUntil)
	return u, err
}
func (r Repository) RevokeRefresh(ctx context.Context, userID uint64, raw string) error {
	h := sha256.Sum256([]byte(raw))
	_, err := r.DB.ExecContext(ctx, `UPDATE user_refresh_tokens SET revoked_at=UTC_TIMESTAMP() WHERE user_id=? AND token_hash=?`, userID, hex.EncodeToString(h[:]))
	return err
}
func IsNotFound(err error) bool { return err == sql.ErrNoRows }
