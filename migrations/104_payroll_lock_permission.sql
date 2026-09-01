-- Locking is the terminal payroll workflow action and is permissioned separately
-- from finalization so a company can require tighter control over completed payrolls.

INSERT INTO permissions (code,name,description) VALUES
('payroll.lock','Lock payroll','Lock a finalized payroll run as a completed historical record.')
ON DUPLICATE KEY UPDATE name=VALUES(name),description=VALUES(description);

INSERT IGNORE INTO company_role_template_permissions (role_template_id,permission_id)
SELECT rt.id,p.id
FROM company_role_templates rt
JOIN permissions p ON p.code='payroll.lock'
WHERE rt.code IN ('OWNER','ADMIN');

INSERT IGNORE INTO company_role_permissions (role_id,permission_id)
SELECT cr.id,p.id
FROM company_roles cr
JOIN permissions p ON p.code='payroll.lock'
WHERE cr.code IN ('OWNER','ADMIN')
AND NOT EXISTS (
 SELECT 1 FROM company_role_permissions x
 WHERE x.role_id=cr.id AND x.permission_id=p.id
);