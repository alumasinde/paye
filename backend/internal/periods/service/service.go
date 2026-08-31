package service

import (
	"errors"
	"time"
)

var ErrInvalidPeriod = errors.New("invalid period")

type Period struct {
	Date  time.Time `json:"date"`
	Start time.Time `json:"start"`
	End   time.Time `json:"end"`
	Year  int       `json:"year"`
	Month int       `json:"month"`
}

func Resolve(date time.Time) (Period, error) {
	if date.IsZero() {
		return Period{}, ErrInvalidPeriod
	}
	u := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, time.UTC)
	start := time.Date(u.Year(), u.Month(), 1, 0, 0, 0, 0, time.UTC)
	return Period{Date: u, Start: start, End: start.AddDate(0, 1, -1), Year: u.Year(), Month: int(u.Month())}, nil
}
func Parse(v string) (Period, error) {
	d, e := time.Parse("2006-01-02", v)
	if e != nil {
		return Period{}, ErrInvalidPeriod
	}
	return Resolve(d)
}
