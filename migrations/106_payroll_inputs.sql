-- Phase 3: richer payroll inputs and bulk payroll-wide application.
ALTER TABLE payroll_run_employee_adjustments
 ADD COLUMN category VARCHAR(40) NOT NULL DEFAULT 'OTHER' AFTER kind,
 ADD COLUMN payee VARCHAR(150) NULL AFTER category,
 ADD COLUMN reference_no VARCHAR(100) NULL AFTER payee,
 ADD COLUMN source VARCHAR(30) NOT NULL DEFAULT 'MANUAL' AFTER reference_no;

CREATE INDEX idx_payroll_adjustment_category ON payroll_run_employee_adjustments(category);
