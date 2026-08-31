package model
import "time"

type User struct {
	ID               uint64
	PublicID         string
	Email            string
	PasswordHash     string
	FirstName        string
	LastName         string
	Status           string
	CreatedAt        time.Time
	FailedLoginCount uint
	LockedUntil      *time.Time
}
