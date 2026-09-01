-- Payroll run review, approval, finalization and locking workflow.
-- The workflow is intentionally state-driven and immutable once finalized.

ALTER TABLE payroll_runs
 ADD COLUMN reviewed_by BIGINT UNSIGNED NULL AFTER status,
 ADD COLUMN reviewed_at DATETIME NULL AFTER reviewed_by,
 ADD COLUMN approved_by BIGINT UNSIGNED NULL AFTER reviewed_at,
 ADD COLUMN approved_at DATETIME NULL AFTER approved_by,
 ADD COLUMN finalized_by BIGINT UNSIGNED NULL AFTER approved_at,
 ADD COLUMN finalized_at DATETIME NULL AFTER finalized_by,
 ADD COLUMN locked_by BIGINT UNSIGNED NULL AFTER finalized_at,
 ADD COLUMN locked_at DATETIME NULL AFTER locked_by,
 ADD CONSTRAINT fk_payroll_runs_reviewed_by FOREIGN KEY (reviewed_by) REFERENCES users(id),
 ADD CONSTRAINT fk_payroll_runs_approved_by FOREIGN KEY (approved_by) REFERENCES users(id),
 ADD CONSTRAINT fk_payroll_runs_finalized_by FOREIGN KEY (finalized_by) REFERENCES users(id),
 ADD CONSTRAINT fk_payroll_runs_locked_by FOREIGN KEY (locked_by) REFERENCES users(id);

CREATE INDEX idx_payroll_runs_company_status ON payroll_runs(company_id,status);

INSERT INTO permissions (code,name,description) VALUES
('payroll.review','Review payroll','Move a calculated payroll run into review.'),
('payroll.approve','Approve payroll','Approve a reviewed payroll run.'),
('payroll.finalize','Finalize payroll','Finalize and lock approved payroll results.')
ON DUPLICATE KEY UPDATE name=VALUES(name),description=VALUES(description);

INSERT IGNORE INTO company_role_template_permissions (role_template_id,permission_id)
SELECT rt.id,p.id
FROM company_role_templates rt
JOIN permissions p
WHERE rt.code IN ('OWNER','ADMIN')
  AND p.code IN ('payroll.review','payroll.approve','payroll.finalize');

INSERT IGNORE INTO company_role_template_permissions (role_template_id,permission_id)
SELECT rt.id,p.id
FROM company_role_templates rt
JOIN permissions p
WHERE rt.code='PAYROLL_MANAGER'
  AND p.code IN ('payroll.review','payroll.approve');

INSERT IGNORE INTO company_role_permissions (role_id,permission_id)
SELECT cr.id,p.id
FROM company_roles cr
JOIN permissions p ON p.code IN ('payroll.review','payroll.approve','payroll.finalize')
WHERE (
 (cr.code IN ('OWNER','ADMIN'))
 OR (cr.code='PAYROLL_MANAGER' AND p.code IN ('payroll.review','payroll.approve'))
)
AND NOT EXISTS (
 SELECT 1 FROM company_role_permissions x
 WHERE x.role_id=cr.id AND x.permission_id=p.id
);
