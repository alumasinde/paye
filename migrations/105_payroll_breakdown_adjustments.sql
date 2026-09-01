-- Payroll breakdown and adjustments are stored against the payroll employee
-- snapshot, not the live employee profile. This keeps historical payrolls
-- reproducible and allows one-off earnings and deductions for a specific run.

CREATE TABLE IF NOT EXISTS payroll_run_employee_adjustments (
 id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
 public_id CHAR(36) NOT NULL,
 payroll_run_employee_id BIGINT UNSIGNED NOT NULL,
 name VARCHAR(150) NOT NULL,
 kind VARCHAR(20) NOT NULL,
 amount DECIMAL(18,2) NOT NULL,
 taxable TINYINT(1) NOT NULL DEFAULT 1,
 reduces_taxable_income TINYINT(1) NOT NULL DEFAULT 0,
 created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
 updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
 PRIMARY KEY (id),
 UNIQUE KEY uq_payroll_run_employee_adjustments_public_id (public_id),
 KEY idx_payroll_run_employee_adjustments_employee (payroll_run_employee_id),
 CONSTRAINT fk_payroll_run_employee_adjustments_employee
  FOREIGN KEY (payroll_run_employee_id) REFERENCES payroll_run_employees(id) ON DELETE CASCADE,
 CONSTRAINT chk_payroll_adjustment_kind CHECK (kind IN ('EARNING','DEDUCTION')),
 CONSTRAINT chk_payroll_adjustment_amount CHECK (amount > 0)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;