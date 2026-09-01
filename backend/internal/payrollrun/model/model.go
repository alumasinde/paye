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
}

type Detail struct {
 PayrollRun
 Employees []PayrollRunEmployee `json:"employees"`
}

type CreateInput struct { Period string `json:"period"` }

type CalculationSummary struct {
 PayrollRun
 Processed int `json:"processed"`
 Failed int `json:"failed"`
 Pending int `json:"pending"`
}
