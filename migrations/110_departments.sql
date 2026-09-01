-- Company departments. Employee department text remains temporarily for safe legacy migration.
CREATE TABLE departments (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    public_id CHAR(36) NOT NULL,
    company_id BIGINT UNSIGNED NOT NULL,
    name VARCHAR(150) NOT NULL,
    code VARCHAR(50) NULL,
    description TEXT NULL,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    UNIQUE KEY uq_departments_public_id (public_id),
    UNIQUE KEY uq_departments_company_name (company_id, name),
    UNIQUE KEY uq_departments_company_code (company_id, code),
    KEY idx_departments_company_active (company_id, is_active),
    CONSTRAINT fk_departments_company FOREIGN KEY (company_id) REFERENCES companies(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

ALTER TABLE employees ADD COLUMN department_id BIGINT UNSIGNED NULL AFTER department;
ALTER TABLE employees ADD KEY idx_employees_department (department_id);
ALTER TABLE employees ADD CONSTRAINT fk_employees_department FOREIGN KEY (department_id) REFERENCES departments(id) ON DELETE SET NULL;

-- Preserve existing installations by converting existing non-empty department names
-- into company-specific department records, then linking employees.
INSERT IGNORE INTO departments(public_id, company_id, name)
SELECT UUID(), company_id, TRIM(department)
FROM employees
WHERE department IS NOT NULL AND TRIM(department) <> '';

UPDATE employees e
JOIN departments d ON d.company_id=e.company_id AND d.name=TRIM(e.department)
SET e.department_id=d.id
WHERE e.department_id IS NULL AND e.department IS NOT NULL AND TRIM(e.department) <> '';

INSERT INTO permissions(code,name,description) VALUES
('departments.read','View departments','View company departments.'),
('departments.write','Manage departments','Create and update company departments.')
ON DUPLICATE KEY UPDATE name=VALUES(name), description=VALUES(description);

INSERT IGNORE INTO company_role_template_permissions(role_template_id,permission_id)
SELECT rt.id,p.id FROM company_role_templates rt JOIN permissions p
WHERE rt.code IN ('OWNER','ADMIN','PAYROLL_MANAGER') AND p.code IN ('departments.read','departments.write');

INSERT IGNORE INTO company_role_permissions(role_id,permission_id)
SELECT cr.id,p.id FROM company_roles cr
JOIN company_role_templates rt ON rt.code=cr.code
JOIN permissions p ON p.code IN ('departments.read','departments.write')
WHERE rt.code IN ('OWNER','ADMIN','PAYROLL_MANAGER')
AND NOT EXISTS (SELECT 1 FROM company_role_permissions existing WHERE existing.role_id=cr.id AND existing.permission_id=p.id);