CREATE TABLE IF NOT EXISTS payroll_runs (
 id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
 public_id CHAR(36) NOT NULL,
 company_id BIGINT UNSIGNED NOT NULL,
 period_key CHAR(7) NOT NULL,
 period_start DATE NOT NULL,
 period_end DATE NOT NULL,
 status VARCHAR(30) NOT NULL,
 created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
 updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
 PRIMARY KEY (id),
 UNIQUE KEY uq_payroll_runs_public_id (public_id),
 UNIQUE KEY uq_payroll_runs_company_period (company_id, period_key),
 CONSTRAINT fk_payroll_runs_company FOREIGN KEY (company_id) REFERENCES companies(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS payroll_run_employees (
 id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
 public_id CHAR(36) NOT NULL,
 payroll_run_id BIGINT UNSIGNED NOT NULL,
 employee_id BIGINT UNSIGNED NOT NULL,
 employee_number VARCHAR(100) NOT NULL,
 first_name VARCHAR(100) NOT NULL,
 middle_name VARCHAR(100) NULL,
 last_name VARCHAR(100) NOT NULL,
 basic_salary DECIMAL(18,2) NOT NULL,
 pay_frequency VARCHAR(30) NOT NULL,
 status VARCHAR(30) NOT NULL,
 created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
 updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
 PRIMARY KEY (id),
 UNIQUE KEY uq_payroll_run_employee_public_id (public_id),
 UNIQUE KEY uq_payroll_run_employee (payroll_run_id, employee_id),
 KEY idx_payroll_run_employees_run (payroll_run_id),
 CONSTRAINT fk_payroll_run_employees_run FOREIGN KEY (payroll_run_id) REFERENCES payroll_runs(id) ON DELETE CASCADE,
 CONSTRAINT fk_payroll_run_employees_employee FOREIGN KEY (employee_id) REFERENCES employees(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
