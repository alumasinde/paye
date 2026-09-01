-- Immutable audit history for payroll workflow transitions.
-- Payroll runs are financial records; each review, approval, finalization and
-- lock transition is retained independently from the current run status.

CREATE TABLE IF NOT EXISTS payroll_run_workflow_events (
 id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
 payroll_run_id BIGINT UNSIGNED NOT NULL,
 actor_user_id BIGINT UNSIGNED NOT NULL,
 action VARCHAR(30) NOT NULL,
 from_status VARCHAR(30) NOT NULL,
 to_status VARCHAR(30) NOT NULL,
 created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
 PRIMARY KEY (id),
 KEY idx_payroll_run_workflow_events_run (payroll_run_id, created_at),
 KEY idx_payroll_run_workflow_events_actor (actor_user_id),
 CONSTRAINT fk_payroll_run_workflow_events_run
  FOREIGN KEY (payroll_run_id) REFERENCES payroll_runs(id) ON DELETE CASCADE,
 CONSTRAINT fk_payroll_run_workflow_events_actor
  FOREIGN KEY (actor_user_id) REFERENCES users(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
