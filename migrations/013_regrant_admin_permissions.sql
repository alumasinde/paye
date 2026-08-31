-- cmd/migrate tracks applied migrations by filename only, not by content.
-- The role-permission grants below were added to migration 007's file
-- *after* some databases had already recorded 007 as applied - on any
-- such database, `migrate up` sees "007 already applied" and silently
-- never re-runs it, so the grants never actually landed. Symptom: every
-- admin account gets 403 forbidden on every permission-gated endpoint
-- (rule-sets, audit-logs, users), even SUPER_ADMIN, because
-- admin_role_permissions has no rows at all.
--
-- This migration re-issues those exact grants under a new filename, so
-- it always runs regardless of whichever version of 007 a given database
-- happened to apply. INSERT IGNORE makes it safe to run even on a
-- database that already has the grants (from a fresh 007).

INSERT IGNORE INTO admin_role_permissions(role_id, permission_id)
SELECT ar.id, ap.id FROM admin_roles ar JOIN admin_permissions ap
WHERE ar.code = 'SUPER_ADMIN';

INSERT IGNORE INTO admin_role_permissions(role_id, permission_id)
SELECT ar.id, ap.id FROM admin_roles ar JOIN admin_permissions ap
WHERE ar.code = 'RULE_EDITOR' AND ap.code IN ('rules.read','rules.write');

INSERT IGNORE INTO admin_role_permissions(role_id, permission_id)
SELECT ar.id, ap.id FROM admin_roles ar JOIN admin_permissions ap
WHERE ar.code = 'RULE_APPROVER' AND ap.code IN ('rules.read','rules.review','rules.publish');

INSERT IGNORE INTO admin_role_permissions(role_id, permission_id)
SELECT ar.id, ap.id FROM admin_roles ar JOIN admin_permissions ap
WHERE ar.code = 'AUDITOR' AND ap.code IN ('rules.read','audit.read');
