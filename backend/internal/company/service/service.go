package service

import (
	"context"
	"errors"
	"regexp"
	"strings"

	"github.com/alumasinde/budget254-paye-api/internal/company/model"
	repo "github.com/alumasinde/budget254-paye-api/internal/company/repository"
)

var (
	codePattern  = regexp.MustCompile(`^[A-Z0-9_-]+$`)
	colorPattern = regexp.MustCompile(`^#[0-9A-F]{6}$`)
)

type Service struct{ Repo repo.Repository }

func (s Service) CreateCompany(ctx context.Context, userID string, in model.CreateCompanyInput) (model.Company, error) {
	in = normalizeCreate(in)
	if err := validateCreate(in); err != nil { return model.Company{}, err }
	return s.Repo.CreateCompany(ctx, userID, in)
}

func (s Service) ListCompanies(ctx context.Context, userID string) ([]model.Company, error) {
	return s.Repo.ListCompanies(ctx, userID)
}

func (s Service) Company(ctx context.Context, companyID, userID string) (model.Company, error) {
	return s.Repo.Company(ctx, companyID, userID)
}

func (s Service) UpdateCompany(ctx context.Context, companyID, userID string, in model.UpdateCompanyInput) (model.Company, error) {
	in = normalizeUpdate(in)
	if err := validateUpdate(in); err != nil { return model.Company{}, err }
	ok, err := s.Repo.HasPermission(ctx, companyID, userID, "company.write")
	if err != nil { return model.Company{}, err }
	if !ok { return model.Company{}, repo.ErrForbidden }
	return s.Repo.UpdateCompany(ctx, companyID, in)
}

func (s Service) ListRoles(ctx context.Context, companyID, userID string) ([]model.Role, error) {
	if err := s.require(ctx, companyID, userID, "roles.read"); err != nil { return nil, err }
	return s.Repo.ListRoles(ctx, companyID)
}

func (s Service) CreateRole(ctx context.Context, companyID, userID string, in model.CreateRoleInput) (model.Role, error) {
	if err := s.require(ctx, companyID, userID, "roles.write"); err != nil { return model.Role{}, err }
	in.Code = strings.ToUpper(strings.TrimSpace(in.Code))
	in.Name = strings.TrimSpace(in.Name)
	in.Description = strings.TrimSpace(in.Description)
	if in.Code == "" || in.Name == "" { return model.Role{}, errors.New("role code and name are required") }
	if !codePattern.MatchString(in.Code) { return model.Role{}, errors.New("role code may contain only letters, numbers, underscores and hyphens") }
	return s.Repo.CreateRole(ctx, companyID, in)
}

func (s Service) ListMembers(ctx context.Context, companyID, userID string) ([]model.Member, error) {
	if err := s.require(ctx, companyID, userID, "members.read"); err != nil { return nil, err }
	return s.Repo.ListMembers(ctx, companyID)
}

func (s Service) AddMember(ctx context.Context, companyID, userID, email, roleID string) (model.Member, error) {
	if err := s.require(ctx, companyID, userID, "members.write"); err != nil { return model.Member{}, err }
	email = strings.ToLower(strings.TrimSpace(email))
	roleID = strings.TrimSpace(roleID)
	if email == "" || roleID == "" { return model.Member{}, errors.New("email and role_id are required") }
	return s.Repo.AddMember(ctx, companyID, email, roleID)
}

func (s Service) UpdateMemberRole(ctx context.Context, companyID, userID, memberID, roleID string) (model.Member, error) {
	if err := s.require(ctx, companyID, userID, "members.write"); err != nil { return model.Member{}, err }
	if strings.TrimSpace(memberID) == "" || strings.TrimSpace(roleID) == "" { return model.Member{}, errors.New("member_id and role_id are required") }
	return s.Repo.UpdateMemberRole(ctx, companyID, memberID, roleID)
}

func (s Service) require(ctx context.Context, companyID, userID, permission string) error {
	ok, err := s.Repo.HasPermission(ctx, companyID, userID, permission)
	if err != nil { return err }
	if !ok { return repo.ErrForbidden }
	return nil
}

func normalizeCreate(in model.CreateCompanyInput) model.CreateCompanyInput {
	in.LegalName = strings.TrimSpace(in.LegalName)
	in.TradingName = strings.TrimSpace(in.TradingName)
	in.KRAPIN = strings.ToUpper(strings.TrimSpace(in.KRAPIN))
	in.Email = strings.ToLower(strings.TrimSpace(in.Email))
	in.Phone = strings.TrimSpace(in.Phone)
	in.CountryCode = strings.ToUpper(strings.TrimSpace(in.CountryCode))
	in.CurrencyCode = strings.ToUpper(strings.TrimSpace(in.CurrencyCode))
	in.PayrollFrequency = strings.ToUpper(strings.TrimSpace(in.PayrollFrequency))
	in.PrimaryColor = normalizeColor(in.PrimaryColor, "#15803D")
	in.SecondaryColor = normalizeColor(in.SecondaryColor, "#166534")
	return in
}

func normalizeUpdate(in model.UpdateCompanyInput) model.UpdateCompanyInput {
	in.LegalName = strings.TrimSpace(in.LegalName)
	in.TradingName = strings.TrimSpace(in.TradingName)
	in.Email = strings.ToLower(strings.TrimSpace(in.Email))
	in.Phone = strings.TrimSpace(in.Phone)
	in.CountryCode = strings.ToUpper(strings.TrimSpace(in.CountryCode))
	in.CurrencyCode = strings.ToUpper(strings.TrimSpace(in.CurrencyCode))
	in.PayrollFrequency = strings.ToUpper(strings.TrimSpace(in.PayrollFrequency))
	in.PrimaryColor = normalizeColor(in.PrimaryColor, "#15803D")
	in.SecondaryColor = normalizeColor(in.SecondaryColor, "#166534")
	return in
}

func validateCreate(in model.CreateCompanyInput) error {
	if in.LegalName == "" || in.KRAPIN == "" || in.Email == "" { return errors.New("legal_name, kra_pin and email are required") }
	if err := validateShared(in.CountryCode, in.CurrencyCode, in.PayrollFrequency); err != nil { return err }
	return validateColors(in.PrimaryColor, in.SecondaryColor)
}

func validateUpdate(in model.UpdateCompanyInput) error {
	if in.LegalName == "" || in.Email == "" { return errors.New("legal_name and email are required") }
	if err := validateShared(in.CountryCode, in.CurrencyCode, in.PayrollFrequency); err != nil { return err }
	return validateColors(in.PrimaryColor, in.SecondaryColor)
}

func validateShared(country, currency, frequency string) error {
	if len(country) != 2 || !codePattern.MatchString(country) { return errors.New("country_code must be a two-character code") }
	if len(currency) != 3 || !codePattern.MatchString(currency) { return errors.New("currency_code must be a three-character code") }
	if frequency == "" || len(frequency) > 30 || !codePattern.MatchString(frequency) { return errors.New("payroll_frequency must be a valid code") }
	return nil
}

func normalizeColor(value, fallback string) string {
	value = strings.ToUpper(strings.TrimSpace(value))
	if value == "" { return fallback }
	if !strings.HasPrefix(value, "#") { value = "#" + value }
	return value
}

func validateColors(primary, secondary string) error {
	if !colorPattern.MatchString(primary) || !colorPattern.MatchString(secondary) {
		return errors.New("primary_color and secondary_color must be valid hex colors")
	}
	return nil
}
