-- Employee foundation for multi-company payroll.
-- Configurable values are stored as VARCHAR values, not database ENUMs.

CREATE TABLE employees (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    public_id CHAR(36) NOT NULL,
    company_id BIGINT UNSIGNED NOT NULL,
    employee_number VARCHAR(100) NOT NULL,
    first_name VARCHAR(100) NOT NULL,
    middle_name VARCHAR(100) NULL,
    last_name VARCHAR(100) NOT NULL,
    gender ENUM('MALE','FEMALE','OTHER','PREFER_NOT_TO_SAY') NULL,
    date_of_birth DATE NULL,
    national_id VARCHAR(100) NULL,
    passport_number VARCHAR(100) NULL,
    nationality VARCHAR(100) NULL,
    kra_pin VARCHAR(50) NULL,
    nssf_number VARCHAR(100) NULL,
    shif_number VARCHAR(100) NULL,
    nhif_number VARCHAR(100) NULL,
    employment_date DATE NOT NULL,
    termination_date DATE NULL,
    job_title VARCHAR(150) NULL,
    department VARCHAR(150) NULL,
    employment_type VARCHAR(100) NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'ACTIVE',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    UNIQUE KEY uq_employees_public_id (public_id),
    UNIQUE KEY uq_employees_company_number (company_id, employee_number),
    KEY idx_employees_company_status (company_id, status),
    CONSTRAINT fk_employees_company FOREIGN KEY (company_id) REFERENCES companies(id) ON DELETE CASCADE,
    CONSTRAINT chk_employee_termination CHECK (termination_date IS NULL OR termination_date >= employment_date)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE employee_salary_history (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    public_id CHAR(36) NOT NULL,
    employee_id BIGINT UNSIGNED NOT NULL,
    basic_salary DECIMAL(18,2) NOT NULL,
    pay_frequency VARCHAR(30) NOT NULL,
    effective_from DATE NOT NULL,
    effective_to DATE NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    UNIQUE KEY uq_employee_salary_history_public_id (public_id),
    UNIQUE KEY uq_employee_salary_effective_from (employee_id, effective_from),
    KEY idx_employee_salary_current (employee_id, effective_from, effective_to),
    CONSTRAINT fk_employee_salary_history_employee FOREIGN KEY (employee_id) REFERENCES employees(id) ON DELETE CASCADE,
    CONSTRAINT chk_employee_salary_amount CHECK (basic_salary >= 0),
    CONSTRAINT chk_employee_salary_dates CHECK (effective_to IS NULL OR effective_to >= effective_from)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

INSERT INTO permissions (code, name, description) VALUES
('employees.read', 'View employees', 'View employees and salary history.'),
('employees.write', 'Manage employees', 'Create and update employees and salary history.')
ON DUPLICATE KEY UPDATE name=VALUES(name), description=VALUES(description);

INSERT IGNORE INTO company_role_template_permissions (role_template_id, permission_id)
SELECT rt.id, p.id
FROM company_role_templates rt
JOIN permissions p
WHERE rt.code IN ('OWNER','ADMIN','PAYROLL_MANAGER')
  AND p.code IN ('employees.read','employees.write');

INSERT IGNORE INTO company_role_permissions (role_id, permission_id)
SELECT cr.id, p.id
FROM company_roles cr
JOIN company_role_templates rt ON rt.code=cr.code
JOIN permissions p ON p.code IN ('employees.read','employees.write')
WHERE rt.code IN ('OWNER','ADMIN','PAYROLL_MANAGER')
  AND NOT EXISTS (
      SELECT 1 FROM company_role_permissions existing
      WHERE existing.role_id=cr.id AND existing.permission_id=p.id
  );