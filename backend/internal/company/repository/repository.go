package repository

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"github.com/google/uuid"

	"github.com/alumasinde/budget254-paye-api/internal/company/model"
)

var (
	ErrNotFound   = errors.New("not found")
	ErrForbidden  = errors.New("forbidden")
	ErrConflict   = errors.New("conflict")
)

type Repository struct{ DB *sql.DB }

func (r Repository) CreateCompany(ctx context.Context, userPublicID string, in model.CreateCompanyInput) (model.Company, error) {
	tx, err := r.DB.BeginTx(ctx, nil)
	if err != nil {
		return model.Company{}, err
	}
	defer tx.Rollback()

	var userID uint64
	if err := tx.QueryRowContext(ctx, `SELECT id FROM users WHERE public_id=? AND status='ACTIVE'`, userPublicID).Scan(&userID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.Company{}, ErrNotFound
		}
		return model.Company{}, err
	}

	publicID := uuid.NewString()
	_, err = tx.ExecContext(ctx, `
		INSERT INTO companies
		(public_id,legal_name,trading_name,kra_pin,email,phone,country_code,currency_code,payroll_frequency,created_by)
		VALUES (?,?,?,?,?,?,?,?,?,?)`,
		publicID, in.LegalName, nullable(in.TradingName), in.KRAPIN, in.Email,
		nullable(in.Phone), in.CountryCode, in.CurrencyCode, in.PayrollFrequency, userID,
	)
	if err != nil {
		if isDuplicate(err) {
			return model.Company{}, ErrConflict
		}
		return model.Company{}, err
	}

	var companyID uint64
	if err := tx.QueryRowContext(ctx, `SELECT id FROM companies WHERE public_id=?`, publicID).Scan(&companyID); err != nil {
		return model.Company{}, err
	}

	if _, err := tx.ExecContext(ctx, `
		INSERT INTO company_roles (public_id,company_id,code,name,description,is_system,is_active)
		SELECT UUID(), ?, code, name, description, is_system, 1
		FROM company_role_templates`, companyID); err != nil {
		return model.Company{}, err
	}

	if _, err := tx.ExecContext(ctx, `
		INSERT INTO company_role_permissions (role_id,permission_id)
		SELECT cr.id, rtp.permission_id
		FROM company_roles cr
		JOIN company_role_templates rt ON rt.code=cr.code
		JOIN company_role_template_permissions rtp ON rtp.role_template_id=rt.id
		WHERE cr.company_id=?`, companyID); err != nil {
		return model.Company{}, err
	}

	var ownerRoleID uint64
	if err := tx.QueryRowContext(ctx, `SELECT id FROM company_roles WHERE company_id=? AND code='OWNER'`, companyID).Scan(&ownerRoleID); err != nil {
		return model.Company{}, err
	}

	if _, err := tx.ExecContext(ctx, `
		INSERT INTO company_members (public_id,company_id,user_id,role_id,status)
		VALUES (?,?,?,?, 'ACTIVE')`, uuid.NewString(), companyID, userID, ownerRoleID); err != nil {
		return model.Company{}, err
	}

	if err := tx.Commit(); err != nil {
		return model.Company{}, err
	}
	return r.Company(ctx, publicID, userPublicID)
}

func (r Repository) ListCompanies(ctx context.Context, userPublicID string) ([]model.Company, error) {
	rows, err := r.DB.QueryContext(ctx, `
		SELECT c.public_id,c.legal_name,COALESCE(c.trading_name,''),c.kra_pin,c.email,
		       COALESCE(c.phone,''),c.country_code,c.currency_code,c.payroll_frequency,c.status,c.created_at
		FROM companies c
		JOIN company_members cm ON cm.company_id=c.id AND cm.status='ACTIVE'
		JOIN users u ON u.id=cm.user_id
		WHERE u.public_id=?
		ORDER BY c.created_at ASC`, userPublicID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]model.Company, 0)
	for rows.Next() {
		var c model.Company
		if err := rows.Scan(&c.PublicID,&c.LegalName,&c.TradingName,&c.KRAPIN,&c.Email,&c.Phone,&c.CountryCode,&c.CurrencyCode,&c.PayrollFrequency,&c.Status,&c.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

func (r Repository) Company(ctx context.Context, companyPublicID, userPublicID string) (model.Company, error) {
	var c model.Company
	err := r.DB.QueryRowContext(ctx, `
		SELECT c.public_id,c.legal_name,COALESCE(c.trading_name,''),c.kra_pin,c.email,
		       COALESCE(c.phone,''),c.country_code,c.currency_code,c.payroll_frequency,c.status,c.created_at
		FROM companies c
		JOIN company_members cm ON cm.company_id=c.id AND cm.status='ACTIVE'
		JOIN users u ON u.id=cm.user_id
		WHERE c.public_id=? AND u.public_id=?`, companyPublicID, userPublicID).
		Scan(&c.PublicID,&c.LegalName,&c.TradingName,&c.KRAPIN,&c.Email,&c.Phone,&c.CountryCode,&c.CurrencyCode,&c.PayrollFrequency,&c.Status,&c.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return model.Company{}, ErrNotFound
	}
	return c, err
}

func (r Repository) UpdateCompany(ctx context.Context, companyPublicID string, in model.UpdateCompanyInput) (model.Company, error) {
	res, err := r.DB.ExecContext(ctx, `
		UPDATE companies
		SET legal_name=?,trading_name=?,email=?,phone=?,country_code=?,currency_code=?,payroll_frequency=?
		WHERE public_id=? AND status='ACTIVE'`,
		in.LegalName, nullable(in.TradingName), in.Email, nullable(in.Phone),
		in.CountryCode, in.CurrencyCode, in.PayrollFrequency, companyPublicID,
	)
	if err != nil {
		return model.Company{}, err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return model.Company{}, ErrNotFound
	}
	return r.companyByPublicID(ctx, companyPublicID)
}

func (r Repository) HasPermission(ctx context.Context, companyPublicID, userPublicID, permissionCode string) (bool, error) {
	var n int
	err := r.DB.QueryRowContext(ctx, `
		SELECT COUNT(*)
		FROM company_members cm
		JOIN companies c ON c.id=cm.company_id
		JOIN users u ON u.id=cm.user_id
		JOIN company_role_permissions crp ON crp.role_id=cm.role_id
		JOIN permissions p ON p.id=crp.permission_id
		WHERE c.public_id=? AND u.public_id=? AND cm.status='ACTIVE'
		  AND c.status='ACTIVE' AND p.code=?`,
		companyPublicID, userPublicID, permissionCode,
	).Scan(&n)
	return n > 0, err
}

func (r Repository) ListRoles(ctx context.Context, companyPublicID string) ([]model.Role, error) {
	companyID, err := r.companyID(ctx, companyPublicID)
	if err != nil {
		return nil, err
	}
	rows, err := r.DB.QueryContext(ctx, `
		SELECT public_id,code,name,COALESCE(description,''),is_system,is_active
		FROM company_roles WHERE company_id=? ORDER BY is_system DESC,name ASC`, companyID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]model.Role, 0)
	for rows.Next() {
		var role model.Role
		if err := rows.Scan(&role.PublicID,&role.Code,&role.Name,&role.Description,&role.IsSystem,&role.IsActive); err != nil {
			return nil, err
		}
		role.Permissions, err = r.permissionsForRole(ctx, role.PublicID)
		if err != nil {
			return nil, err
		}
		out = append(out, role)
	}
	return out, rows.Err()
}

func (r Repository) CreateRole(ctx context.Context, companyPublicID string, in model.CreateRoleInput) (model.Role, error) {
	companyID, err := r.companyID(ctx, companyPublicID)
	if err != nil {
		return model.Role{}, err
	}
	tx, err := r.DB.BeginTx(ctx, nil)
	if err != nil {
		return model.Role{}, err
	}
	defer tx.Rollback()
	publicID := uuid.NewString()
	if _, err := tx.ExecContext(ctx, `
		INSERT INTO company_roles(public_id,company_id,code,name,description,is_system,is_active)
		VALUES(?,?,?,?,?,0,1)`, publicID, companyID, in.Code, in.Name, nullable(in.Description)); err != nil {
		if isDuplicate(err) {
			return model.Role{}, ErrConflict
		}
		return model.Role{}, err
	}
	var roleID uint64
	if err := tx.QueryRowContext(ctx, `SELECT id FROM company_roles WHERE public_id=?`, publicID).Scan(&roleID); err != nil {
		return model.Role{}, err
	}
	for _, code := range unique(in.Permissions) {
		res, err := tx.ExecContext(ctx, `
			INSERT INTO company_role_permissions(role_id,permission_id)
			SELECT ?, id FROM permissions WHERE code=?`, roleID, code)
		if err != nil {
			return model.Role{}, err
		}
		n, _ := res.RowsAffected()
		if n == 0 {
			return model.Role{}, ErrNotFound
		}
	}
	if err := tx.Commit(); err != nil {
		return model.Role{}, err
	}
	return r.role(ctx, publicID)
}

func (r Repository) ListMembers(ctx context.Context, companyPublicID string) ([]model.Member, error) {
	companyID, err := r.companyID(ctx, companyPublicID)
	if err != nil {
		return nil, err
	}
	rows, err := r.DB.QueryContext(ctx, `
		SELECT cm.public_id,u.public_id,u.email,COALESCE(u.first_name,''),COALESCE(u.last_name,''),
		       cm.status,cm.joined_at,cr.public_id,cr.code,cr.name,COALESCE(cr.description,''),cr.is_system,cr.is_active
		FROM company_members cm
		JOIN users u ON u.id=cm.user_id
		JOIN company_roles cr ON cr.id=cm.role_id
		WHERE cm.company_id=? AND cm.status <> 'REMOVED'
		ORDER BY cm.joined_at ASC`, companyID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]model.Member, 0)
	for rows.Next() {
		var m model.Member
		if err := rows.Scan(&m.PublicID,&m.UserID,&m.Email,&m.FirstName,&m.LastName,&m.Status,&m.JoinedAt,
			&m.Role.PublicID,&m.Role.Code,&m.Role.Name,&m.Role.Description,&m.Role.IsSystem,&m.Role.IsActive); err != nil {
			return nil, err
		}
		m.Role.Permissions, err = r.permissionsForRole(ctx, m.Role.PublicID)
		if err != nil {
			return nil, err
		}
		out = append(out, m)
	}
	return out, rows.Err()
}

func (r Repository) AddMember(ctx context.Context, companyPublicID, email, rolePublicID string) (model.Member, error) {
	companyID, err := r.companyID(ctx, companyPublicID)
	if err != nil {
		return model.Member{}, err
	}
	var userID uint64
	if err := r.DB.QueryRowContext(ctx, `SELECT id FROM users WHERE email=? AND status='ACTIVE'`, strings.ToLower(strings.TrimSpace(email))).Scan(&userID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.Member{}, ErrNotFound
		}
		return model.Member{}, err
	}
	var roleID uint64
	if err := r.DB.QueryRowContext(ctx, `SELECT id FROM company_roles WHERE public_id=? AND company_id=? AND is_active=1`, rolePublicID, companyID).Scan(&roleID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.Member{}, ErrNotFound
		}
		return model.Member{}, err
	}
	publicID := uuid.NewString()
	_, err = r.DB.ExecContext(ctx, `
		INSERT INTO company_members(public_id,company_id,user_id,role_id,status)
		VALUES(?,?,?,?, 'ACTIVE')
		ON DUPLICATE KEY UPDATE role_id=VALUES(role_id),status='ACTIVE',updated_at=UTC_TIMESTAMP()`,
		publicID, companyID, userID, roleID)
	if err != nil {
		return model.Member{}, err
	}
	return r.memberByCompanyAndUser(ctx, companyID, userID)
}

func (r Repository) UpdateMemberRole(ctx context.Context, companyPublicID, memberPublicID, rolePublicID string) (model.Member, error) {
	companyID, err := r.companyID(ctx, companyPublicID)
	if err != nil {
		return model.Member{}, err
	}
	var roleID uint64
	if err := r.DB.QueryRowContext(ctx, `SELECT id FROM company_roles WHERE public_id=? AND company_id=? AND is_active=1`, rolePublicID, companyID).Scan(&roleID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.Member{}, ErrNotFound
		}
		return model.Member{}, err
	}
	res, err := r.DB.ExecContext(ctx, `UPDATE company_members SET role_id=? WHERE public_id=? AND company_id=? AND status='ACTIVE'`, roleID, memberPublicID, companyID)
	if err != nil {
		return model.Member{}, err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return model.Member{}, ErrNotFound
	}
	var userID uint64
	if err := r.DB.QueryRowContext(ctx, `SELECT user_id FROM company_members WHERE public_id=? AND company_id=?`, memberPublicID, companyID).Scan(&userID); err != nil {
		return model.Member{}, err
	}
	return r.memberByCompanyAndUser(ctx, companyID, userID)
}

func (r Repository) companyID(ctx context.Context, publicID string) (uint64, error) {
	var id uint64
	err := r.DB.QueryRowContext(ctx, `SELECT id FROM companies WHERE public_id=?`, publicID).Scan(&id)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, ErrNotFound
	}
	return id, err
}

func (r Repository) companyByPublicID(ctx context.Context, publicID string) (model.Company, error) {
	var c model.Company
	err := r.DB.QueryRowContext(ctx, `
		SELECT public_id,legal_name,COALESCE(trading_name,''),kra_pin,email,COALESCE(phone,''),
		       country_code,currency_code,payroll_frequency,status,created_at
		FROM companies WHERE public_id=?`, publicID).
		Scan(&c.PublicID,&c.LegalName,&c.TradingName,&c.KRAPIN,&c.Email,&c.Phone,&c.CountryCode,&c.CurrencyCode,&c.PayrollFrequency,&c.Status,&c.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return model.Company{}, ErrNotFound
	}
	return c, err
}

func (r Repository) role(ctx context.Context, publicID string) (model.Role, error) {
	var role model.Role
	err := r.DB.QueryRowContext(ctx, `SELECT public_id,code,name,COALESCE(description,''),is_system,is_active FROM company_roles WHERE public_id=?`, publicID).
		Scan(&role.PublicID,&role.Code,&role.Name,&role.Description,&role.IsSystem,&role.IsActive)
	if errors.Is(err, sql.ErrNoRows) {
		return model.Role{}, ErrNotFound
	}
	if err != nil {
		return model.Role{}, err
	}
	role.Permissions, err = r.permissionsForRole(ctx, publicID)
	return role, err
}

func (r Repository) permissionsForRole(ctx context.Context, rolePublicID string) ([]string, error) {
	rows, err := r.DB.QueryContext(ctx, `
		SELECT p.code FROM company_role_permissions crp
		JOIN company_roles cr ON cr.id=crp.role_id
		JOIN permissions p ON p.id=crp.permission_id
		WHERE cr.public_id=? ORDER BY p.code`, rolePublicID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]string, 0)
	for rows.Next() {
		var code string
		if err := rows.Scan(&code); err != nil {
			return nil, err
		}
		out = append(out, code)
	}
	return out, rows.Err()
}

func (r Repository) memberByCompanyAndUser(ctx context.Context, companyID, userID uint64) (model.Member, error) {
	var m model.Member
	err := r.DB.QueryRowContext(ctx, `
		SELECT cm.public_id,u.public_id,u.email,COALESCE(u.first_name,''),COALESCE(u.last_name,''),
		       cm.status,cm.joined_at,cr.public_id,cr.code,cr.name,COALESCE(cr.description,''),cr.is_system,cr.is_active
		FROM company_members cm
		JOIN users u ON u.id=cm.user_id
		JOIN company_roles cr ON cr.id=cm.role_id
		WHERE cm.company_id=? AND cm.user_id=?`, companyID, userID).
		Scan(&m.PublicID,&m.UserID,&m.Email,&m.FirstName,&m.LastName,&m.Status,&m.JoinedAt,
			&m.Role.PublicID,&m.Role.Code,&m.Role.Name,&m.Role.Description,&m.Role.IsSystem,&m.Role.IsActive)
	if errors.Is(err, sql.ErrNoRows) {
		return model.Member{}, ErrNotFound
	}
	if err != nil {
		return model.Member{}, err
	}
	m.Role.Permissions, err = r.permissionsForRole(ctx, m.Role.PublicID)
	return m, err
}

func nullable(v string) any {
	v = strings.TrimSpace(v)
	if v == "" {
		return nil
	}
	return v
}

func unique(values []string) []string {
	seen := make(map[string]struct{}, len(values))
	out := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		out = append(out, value)
	}
	return out
}

func isDuplicate(err error) bool {
	return err != nil && strings.Contains(strings.ToLower(err.Error()), "duplicate")
}
