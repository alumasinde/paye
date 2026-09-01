ALTER TABLE payroll_run_employees
 ADD COLUMN gross_salary DECIMAL(18,2) NULL AFTER status,
 ADD COLUMN taxable_income DECIMAL(18,2) NULL AFTER gross_salary,
 ADD COLUMN paye_before_relief DECIMAL(18,2) NULL AFTER taxable_income,
 ADD COLUMN relief DECIMAL(18,2) NULL AFTER paye_before_relief,
 ADD COLUMN paye DECIMAL(18,2) NULL AFTER relief,
 ADD COLUMN statutory_deductions DECIMAL(18,2) NULL AFTER paye,
 ADD COLUMN custom_deductions DECIMAL(18,2) NOT NULL DEFAULT 0.00 AFTER statutory_deductions,
 ADD COLUMN total_deductions DECIMAL(18,2) NULL AFTER custom_deductions,
 ADD COLUMN net_salary DECIMAL(18,2) NULL AFTER total_deductions,
 ADD COLUMN rule_versions JSON NULL AFTER net_salary,
 ADD COLUMN calculated_at DATETIME NULL AFTER rule_versions,
 ADD COLUMN error_message VARCHAR(500) NULL AFTER calculated_at;
