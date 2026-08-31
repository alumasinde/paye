package model
import ("time"; "github.com/shopspring/decimal")
type DeductionType string
const ( NetPay DeductionType="NET_PAY"; TaxableIncome DeductionType="TAXABLE_INCOME" )
type CustomDeduction struct { Name string; Amount decimal.Decimal; Type DeductionType }
type Rule struct {
 Code string; Name string; Version string; Method string
 EffectiveFrom time.Time; EffectiveTo *time.Time
 Parameters map[string]decimal.Decimal
 Bands []Band
 ReducesTaxableIncome bool
}
type Band struct { From decimal.Decimal; To *decimal.Decimal; Rate decimal.Decimal; Amount *decimal.Decimal }
type Result struct {
 Date time.Time; Gross, TaxableIncome, PAYEBeforeRelief, Relief, PAYE, TotalDeductions, Net decimal.Decimal
 Statutory []Item; Custom []Item; RuleVersions map[string]string; Trace []TraceBand
}
type Item struct { Code,Name string; Amount decimal.Decimal; ReducesTaxableIncome bool }
type TraceBand struct { From decimal.Decimal; To *decimal.Decimal; Rate,Tax decimal.Decimal }
