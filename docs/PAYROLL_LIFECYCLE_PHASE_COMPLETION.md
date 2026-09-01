# Employee and Payroll Lifecycle Phase Completion

## Scope completed

The PAYE application now supports the complete company payroll lifecycle:

1. Create departments and employee records.
2. Maintain employee salary history with effective dates.
3. Create a monthly payroll run from active employees and their effective salary snapshots.
4. Add payroll-run earnings and deduction adjustments.
5. Calculate PAYE, statutory deductions, taxable income and net salary using the rules effective for the payroll period.
6. Review a fully calculated payroll run.
7. Approve a reviewed payroll run.
8. Finalize an approved payroll run.
9. Lock a finalized payroll run.
10. View payroll summaries and employee payslips.

## Workflow

DRAFT -> CALCULATED -> REVIEW -> APPROVED -> FINALIZED -> LOCKED

A calculation with failed or pending employees is placed in CALCULATION_FAILED and cannot enter review until the payroll is successfully recalculated.

## Integrity rules

- Payroll runs snapshot employee identity, salary and pay frequency for the selected period.
- Only active employees with an effective salary for the payroll period are included.
- Payroll adjustments are editable only while the run is in DRAFT or CALCULATION_FAILED.
- Review requires every payroll employee to be calculated.
- Workflow transitions require explicit company permissions.
- Workflow transitions are retained in payroll_run_workflow_events as an audit trail.
- payroll.lock is granted to OWNER and ADMIN roles so the final lifecycle transition is executable.
- Payslips are available only after the payroll has been finalized or locked.

## Operational acceptance checklist

Before production release, run this workflow against a clean database and a representative company:

- [ ] Apply all migrations in order.
- [ ] Create a company and department.
- [ ] Create an active employee with an effective salary.
- [ ] Create a payroll run.
- [ ] Add an earning and a deduction adjustment.
- [ ] Calculate and verify the employee breakdown.
- [ ] Review the run.
- [ ] Approve the run.
- [ ] Finalize the run.
- [ ] Lock the run.
- [ ] Open the payroll summary and payslip.
- [ ] Verify workflow events and locked status in the database.

## Remaining work outside this phase

This completion does not claim statutory filing, payment submission, bank files, accounting journal export, or automated employee delivery of payslips. Those should be treated as separate future phases.
