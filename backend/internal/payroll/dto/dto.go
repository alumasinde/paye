package dto
type CalculateRequest struct {
 GrossSalary string `json:"gross_salary"`
 CalculationDate string `json:"calculation_date"`
 CustomDeductions []CustomDeductionRequest `json:"custom_deductions,omitempty"`
 Explain bool `json:"explain,omitempty"`
}
type CustomDeductionRequest struct { Name string `json:"name"`; Amount string `json:"amount"`; Type string `json:"type"` }
type Deduction struct { Code string `json:"code"`; Name string `json:"name"`; Amount string `json:"amount"`; TaxableIncomeEffect bool `json:"reduces_taxable_income"` }
type BandTrace struct { From string `json:"from"`; To *string `json:"to,omitempty"`; Rate string `json:"rate"`; Tax string `json:"tax"` }
type CalculateResponse struct {
 CalculationDate string `json:"calculation_date"`
 GrossSalary string `json:"gross_salary"`
 TaxableIncome string `json:"taxable_income"`
 PAYEBeforeRelief string `json:"paye_before_relief"`
 Relief string `json:"relief"`
 PAYE string `json:"paye"`
 StatutoryDeductions []Deduction `json:"statutory_deductions"`
 CustomDeductions []Deduction `json:"custom_deductions"`
 TotalDeductions string `json:"total_deductions"`
 NetSalary string `json:"net_salary"`
 RuleVersions map[string]string `json:"rule_versions"`
 Trace []BandTrace `json:"trace,omitempty"`
}
type RulesResponse struct { CalculationDate string `json:"calculation_date"`; Rules []Rule `json:"rules"` }
type Rule struct { Code string `json:"code"`; Name string `json:"name"`; Version string `json:"version"`; Method string `json:"calculation_method"`; EffectiveFrom string `json:"effective_from"`; EffectiveTo *string `json:"effective_to,omitempty"` }
