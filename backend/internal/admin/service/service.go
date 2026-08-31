package service

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"github.com/alumasinde/budget254-paye-api/internal/admin/model"
	repo "github.com/alumasinde/budget254-paye-api/internal/admin/repository"
)

const lockoutWindow = 15 * time.Minute

var ErrLocked = errors.New("account is temporarily locked after too many failed attempts")
var ErrInvalidCredentials = errors.New("invalid credentials")

type Service struct {
	Repo                  repo.Repository
	Secret                []byte
	AccessTTL, RefreshTTL time.Duration
	MaxFailedLogins       int
}

type Tokens struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int64  `json:"expires_in"`
}

// Login verifies an admin's email/password and issues a scoped JWT. The
// token carries "typ":"admin" plus the admin's roles/permissions at issue
// time, so route-level authorization (internal/middleware.RequireAdmin /
// RequirePermission) never has to hit the database per request. A customer
// (internal/auth) token can never satisfy an admin route: it has no "typ"
// claim at all.
func (s Service) Login(ctx context.Context, email, password, ua string) (model.AdminUser, Tokens, error) {
	u, err := s.Repo.ByEmail(ctx, strings.ToLower(strings.TrimSpace(email)))
	if err != nil {
		return model.AdminUser{}, Tokens{}, ErrInvalidCredentials
	}
	if u.LockedUntil != nil && u.LockedUntil.After(time.Now()) {
		return model.AdminUser{}, Tokens{}, ErrLocked
	}
	if u.Status != "ACTIVE" || bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password)) != nil {
		if s.MaxFailedLogins > 0 {
			var lockUntil *time.Time
			if int(u.FailedLoginCount)+1 >= s.MaxFailedLogins {
				t := time.Now().Add(lockoutWindow)
				lockUntil = &t
			}
			_ = s.Repo.RegisterFailedLogin(ctx, u.ID, lockUntil)
		}
		return model.AdminUser{}, Tokens{}, ErrInvalidCredentials
	}
	_ = s.Repo.ResetFailedLogins(ctx, u.ID)
	t, err := s.issue(ctx, u, ua)
	return u, t, err
}

// Refresh exchanges a valid, unexpired refresh token for a new token pair,
// rotating it (the old one is revoked once the new one is issued).
func (s Service) Refresh(ctx context.Context, rawRefresh, ua string) (model.AdminUser, Tokens, error) {
	u, err := s.Repo.AdminByRefresh(ctx, rawRefresh)
	if err != nil {
		return model.AdminUser{}, Tokens{}, errors.New("invalid or expired refresh token")
	}
	t, err := s.issue(ctx, u, ua)
	if err != nil {
		return model.AdminUser{}, Tokens{}, err
	}
	_ = s.Repo.RevokeRefresh(ctx, u.ID, rawRefresh)
	return u, t, nil
}

func (s Service) ChangePassword(ctx context.Context, adminPublicID, oldPassword, newPassword string) error {
	if len(newPassword) < 12 {
		return errors.New("new password must be at least 12 characters")
	}
	u, err := s.Repo.ByPublicID(ctx, adminPublicID)
	if err != nil {
		return errors.New("account not found")
	}
	if bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(oldPassword)) != nil {
		return ErrInvalidCredentials
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	return s.Repo.UpdatePassword(ctx, u.ID, string(hash))
}

// CreateAdmin creates a new admin account with the given role. Only meant
// to be called from a route gated on the admin.users permission.
func (s Service) CreateAdmin(ctx context.Context, email, password, first, last, roleCode string) (model.AdminUser, error) {
	email = strings.ToLower(strings.TrimSpace(email))
	first, last = strings.TrimSpace(first), strings.TrimSpace(last)
	if email == "" || first == "" || last == "" || roleCode == "" {
		return model.AdminUser{}, errors.New("email, first name, last name and role are required")
	}
	if len(password) < 12 {
		return model.AdminUser{}, errors.New("password must be at least 12 characters")
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return model.AdminUser{}, err
	}
	publicID := uuid.NewString()
	if err := s.Repo.Create(ctx, publicID, email, string(hash), first, last, roleCode); err != nil {
		return model.AdminUser{}, err
	}
	return s.Repo.ByPublicID(ctx, publicID)
}

func (s Service) List(ctx context.Context) ([]model.AdminUser, error) { return s.Repo.List(ctx) }

func (s Service) SetStatus(ctx context.Context, publicID, status string) error {
	if status != "ACTIVE" && status != "DISABLED" {
		return errors.New("status must be ACTIVE or DISABLED")
	}
	return s.Repo.SetStatus(ctx, publicID, status)
}

func (s Service) issue(ctx context.Context, u model.AdminUser, ua string) (Tokens, error) {
	now := time.Now()
	claims := jwt.MapClaims{
		"typ":         "admin",
		"sub":         u.PublicID,
		"uid":         u.ID,
		"email":       u.Email,
		"roles":       u.Roles,
		"permissions": u.Permissions,
		"iat":         now.Unix(),
		"exp":         now.Add(s.AccessTTL).Unix(),
	}
	at, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(s.Secret)
	if err != nil {
		return Tokens{}, err
	}
	rt := newRefreshToken()
	if err = s.Repo.StoreRefresh(ctx, u.ID, rt, now.Add(s.RefreshTTL)); err != nil {
		return Tokens{}, err
	}
	return Tokens{at, rt, "Bearer", int64(s.AccessTTL.Seconds())}, nil
}

func newRefreshToken() string {
	b := make([]byte, 48)
	_, _ = rand.Read(b)
	return base64.RawURLEncoding.EncodeToString(b)
}
