-- Adds failed-login lockout tracking to both account tables, and the
-- rule_versions.source_summary/notes columns already existed for this
-- purpose - only the lockout columns are new here.

ALTER TABLE users
    ADD COLUMN failed_login_count INT UNSIGNED NOT NULL DEFAULT 0,
    ADD COLUMN locked_until DATETIME NULL;

ALTER TABLE admin_users
    ADD COLUMN failed_login_count INT UNSIGNED NOT NULL DEFAULT 0,
    ADD COLUMN locked_until DATETIME NULL;

-- rule_versions.created_by/reviewed_by/approved_by/published_by were
-- defined (migration 001) against `users`, back when that table meant
-- internal rule-governance staff. Migration 006 repurposed `users` for
-- customer accounts instead, and migration 007 introduced `admin_users`
-- as the real admin-identity table - but these four foreign keys were
-- never updated, so they silently still pointed at customer accounts.
-- Publishing a rule set (attributing it to the admin who published it)
-- would fail outright, since the admin's ID only exists in admin_users.
ALTER TABLE rule_versions
    DROP FOREIGN KEY fk_rule_versions_created_by,
    DROP FOREIGN KEY fk_rule_versions_reviewed_by,
    DROP FOREIGN KEY fk_rule_versions_approved_by,
    DROP FOREIGN KEY fk_rule_versions_published_by;

ALTER TABLE rule_versions
    ADD CONSTRAINT fk_rule_versions_created_by FOREIGN KEY (created_by) REFERENCES admin_users(id) ON DELETE SET NULL,
    ADD CONSTRAINT fk_rule_versions_reviewed_by FOREIGN KEY (reviewed_by) REFERENCES admin_users(id) ON DELETE SET NULL,
    ADD CONSTRAINT fk_rule_versions_approved_by FOREIGN KEY (approved_by) REFERENCES admin_users(id) ON DELETE SET NULL,
    ADD CONSTRAINT fk_rule_versions_published_by FOREIGN KEY (published_by) REFERENCES admin_users(id) ON DELETE SET NULL;
