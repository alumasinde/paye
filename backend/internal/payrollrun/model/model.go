package model

import "time"

type PayrollRun struct {
 PublicID string `json:"id"`
 Period string `json:"period"`
 PeriodStart time.Time `json:"period_start"`
 PeriodEnd time.Time `json:"period_end"`
 Status string `json:"status"`
 EmployeeCount int `json:"employee_count"`
 CreatedAt time.Time `json:"created_at"`
 UpdatedAt time.Time `json:"updated_at"`
}

type PayrollRunEmployee struct {
 PublicID string `json:"id"`
 EmployeeID string `json:"employee_id"`
 EmployeeNumber string `json:"employee_number"`
 FirstName string `json:"first_name"`
 MiddleName string `json:"middle_name,omitempty"`
 LastName string `json:"last_name"`
 BasicSalary string `json:"basic_salary"`
 PayFrequency string `json:"pay_frequency"`
 Status string `json:"status"`
 GrossSalary string `json:"gross_salary,omitempty"`
 TaxableIncome string `json:"taxable_income,omitempty"`
 PAYEBeforeRelief string `json:"paye_before_relief,omitempty"`
 Relief string `json:"relief,omitempty"`
 PAYE string `json:"paye,omitempty"`
 StatutoryDeductions string `json:"statutory_deductions,omitempty"`
 CustomDeductions string `json:"custom_deductions,omitempty"`
 TotalDeductions string `json:"total_deductions,omitempty"`
 NetSalary string `json:"net_salary,omitempty"`
 RuleVersions map[string]string `json:"rule_versions,omitempty"`
 CalculatedAt *time.Time `json:"calculated_at,omitempty"`
 ErrorMessage string `json:"error_message,omitempty"`
 Adjustments []PayrollAdjustment `json:"adjustments,omitempty"`
}

type PayrollAdjustment struct {
 PublicID string `json:"id"`
 Name string `json:"name"`
 Kind string `json:"kind"`
 Category string `json:"category"`
 Payee string `json:"payee,omitempty"`
 ReferenceNo string `json:"reference_no,omitempty"`
 Source string `json:"source"`
 Amount string `json:"amount"`
 Taxable bool `json:"taxable"`
 ReducesTaxableIncome bool `json:"reduces_taxable_income"`
 CreatedAt time.Time `json:"created_at"`
 UpdatedAt time.Time `json:"updated_at"`
}

type AdjustmentInput struct {
 Name string `json:"name"`
 Kind string `json:"kind"`
 Category string `json:"category"`
 Payee string `json:"payee"`
 ReferenceNo string `json:"reference_no"`
 Amount string `json:"amount"`
 Taxable bool `json:"taxable"`
 ReducesTaxableIncome bool `json:"reduces_taxable_income"`
}

type WorkflowEvent struct {
 Action string `json:"action"`
 FromStatus string `json:"from_status"`
 ToStatus string `json:"to_status"`
 ActorName string `json:"actor_name,omitempty"`
 CreatedAt time.Time `json:"created_at"`
}

type Detail struct {
 PayrollRun
 Employees []PayrollRunEmployee `json:"employees"`
 Workflow []WorkflowEvent `json:"workflow"`
}

type CreateInput struct { Period string `json:"period"` }

type WorkflowAction struct {
 Action string `json:"action"`
}

type WorkflowSummary struct {
 PayrollRun
 Action string `json:"action"`
}

type CalculationSummary struct {
 PayrollRun
 Processed int `json:"processed"`
 Failed int `json:"failed"`
 Pending int `json:"pending"`
}

type ValidationCheck struct { Severity string `json:"severity"`; EmployeeID string `json:"employee_id,omitempty"`; EmployeeName string `json:"employee_name,omitempty"`; Code string `json:"code"`; Message string `json:"message"` }
type ValidationSummary struct { PayrollRun; Blocking int `json:"blocking"`; Warnings int `json:"warnings"`; Checks []ValidationCheck `json:"checks"` }

type BulkInput struct { Name string `json:"name"`; Kind string `json:"kind"`; Category string `json:"category"`; Payee string `json:"payee"`; ReferenceNo string `json:"reference_no"`; Amount string `json:"amount"`; Taxable bool `json:"taxable"`; ReducesTaxableIncome bool `json:"reduces_taxable_income"`; EmployeeIDs []string `json:"employee_ids"` }
type BulkInputResult struct { Applied int `json:"applied"`; Skipped int `json:"skipped"`; PayrollRunID string `json:"payroll_run_id"` }
