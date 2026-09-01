-- Complete the payroll locking workflow by granting the lock permission.
-- Migration 102 introduced the LOCK transition and migration 103 records it,
-- but no payroll.lock permission was seeded, leaving every user unable to
-- perform the final lifecycle transition.

INSERT INTO permissions (code,name,description) VALUES
('payroll.lock','Lock payroll','Lock a finalized payroll run and prevent further workflow changes.')
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
