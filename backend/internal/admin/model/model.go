package model

import "time"

type AdminUser struct {
	ID               uint64
	PublicID         string
	Email            string
	PasswordHash     string
	FirstName        string
	LastName         string
	Status           string
	Roles            []string
	Permissions      []string
	CreatedAt        time.Time
	FailedLoginCount uint
	LockedUntil      *time.Time
}
