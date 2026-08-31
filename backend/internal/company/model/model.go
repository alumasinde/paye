package model

import "time"

type Company struct {
	PublicID         string    `json:"id"`
	LegalName       string    `json:"legal_name"`
	TradingName     string    `json:"trading_name,omitempty"`
	KRAPIN          string    `json:"kra_pin"`
	Email           string    `json:"email"`
	Phone           string    `json:"phone,omitempty"`
	CountryCode     string    `json:"country_code"`
	CurrencyCode    string    `json:"currency_code"`
	PayrollFrequency string   `json:"payroll_frequency"`
	Status          string    `json:"status"`
	CreatedAt       time.Time `json:"created_at"`
}

type Role struct {
	PublicID    string   `json:"id"`
	Code        string   `json:"code"`
	Name        string   `json:"name"`
	Description string   `json:"description,omitempty"`
	IsSystem    bool     `json:"is_system"`
	IsActive    bool     `json:"is_active"`
	Permissions []string `json:"permissions"`
}

type Member struct {
	PublicID   string    `json:"id"`
	UserID     string    `json:"user_id"`
	Email      string    `json:"email"`
	FirstName  string    `json:"first_name"`
	LastName   string    `json:"last_name"`
	Role       Role      `json:"role"`
	Status     string    `json:"status"`
	JoinedAt   time.Time `json:"joined_at"`
}

type CreateCompanyInput struct {
	LegalName        string
	TradingName      string
	KRAPIN           string
	Email            string
	Phone            string
	CountryCode      string
	CurrencyCode     string
	PayrollFrequency string
}

type UpdateCompanyInput struct {
	LegalName        string
	TradingName      string
	Email            string
	Phone            string
	CountryCode      string
	CurrencyCode     string
	PayrollFrequency string
}

type CreateRoleInput struct {
	Code        string
	Name        string
	Description string
	Permissions []string
}
