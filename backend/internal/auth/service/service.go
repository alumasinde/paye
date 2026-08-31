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

	"github.com/alumasinde/budget254-paye-api/internal/auth/model"
	repo "github.com/alumasinde/budget254-paye-api/internal/auth/repository"
)

const lockoutWindow = 15 * time.Minute

var ErrLocked = errors.New("account is temporarily locked after too many failed attempts")
var ErrInvalidCredentials = errors.New("invalid credentials")

type Service struct {
	Repo                repo.Repository
	Secret              []byte
	AccessTTL, RefreshTTL time.Duration
	MaxFailedLogins     int
}

type Tokens struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int64  `json:"expires_in"`
}

func (s Service) Register(ctx context.Context, email, password, first, last, ua string) (model.User, Tokens, error) {
	email = strings.ToLower(strings.TrimSpace(email))
	first = strings.TrimSpace(first)
	last = strings.TrimSpace(last)
	if email == "" || first == "" || last == "" {
		return model.User{}, Tokens{}, errors.New("required fields missing")
	}
	if len(password) < 12 {
		return model.User{}, Tokens{}, errors.New("password must be at least 12 characters")
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return model.User{}, Tokens{}, err
	}
	u := model.User{PublicID: uuid.NewString(), Email: email, PasswordHash: string(hash), FirstName: first, LastName: last}
	if err = s.Repo.CreateUser(ctx, u); err != nil {
		return model.User{}, Tokens{}, err
	}
	u, err = s.Repo.ByEmail(ctx, email)
	if err != nil {
		return model.User{}, Tokens{}, err
	}
	t, err := s.issue(ctx, u, ua)
	return u, t, err
}

func (s Service) Login(ctx context.Context, email, password, ua string) (model.User, Tokens, error) {
	u, err := s.Repo.ByEmail(ctx, strings.ToLower(strings.TrimSpace(email)))
	if err != nil {
		return model.User{}, Tokens{}, ErrInvalidCredentials
	}
	if u.LockedUntil != nil && u.LockedUntil.After(time.Now()) {
		return model.User{}, Tokens{}, ErrLocked
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
		return model.User{}, Tokens{}, ErrInvalidCredentials
	}
	_ = s.Repo.ResetFailedLogins(ctx, u.ID)
	t, err := s.issue(ctx, u, ua)
	return u, t, err
}

// Refresh exchanges a valid, unexpired refresh token for a new token pair,
// rotating the refresh token (the old one is revoked once the new one is
// issued) so a leaked refresh token has a single use window.
func (s Service) Refresh(ctx context.Context, rawRefresh, ua string) (model.User, Tokens, error) {
	u, err := s.Repo.UserByRefresh(ctx, rawRefresh)
	if err != nil {
		return model.User{}, Tokens{}, errors.New("invalid or expired refresh token")
	}
	t, err := s.issue(ctx, u, ua)
	if err != nil {
		return model.User{}, Tokens{}, err
	}
	_ = s.Repo.RevokeRefresh(ctx, u.ID, rawRefresh)
	return u, t, nil
}

func (s Service) ChangePassword(ctx context.Context, userPublicID, oldPassword, newPassword string) error {
	if len(newPassword) < 12 {
		return errors.New("new password must be at least 12 characters")
	}
	u, err := s.Repo.ByPublicID(ctx, userPublicID)
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

func (s Service) issue(ctx context.Context, u model.User, ua string) (Tokens, error) {
	now := time.Now()
	claims := jwt.MapClaims{"typ": "customer", "sub": u.PublicID, "uid": u.ID, "email": u.Email, "iat": now.Unix(), "exp": now.Add(s.AccessTTL).Unix()}
	at, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(s.Secret)
	if err != nil {
		return Tokens{}, err
	}
	rt := newRefreshToken()
	if err = s.Repo.StoreRefresh(ctx, u.ID, rt, now.Add(s.RefreshTTL), ua); err != nil {
		return Tokens{}, err
	}
	return Tokens{at, rt, "Bearer", int64(s.AccessTTL.Seconds())}, nil
}

func newRefreshToken() string {
	b := make([]byte, 48)
	_, _ = rand.Read(b)
	return base64.RawURLEncoding.EncodeToString(b)
}
