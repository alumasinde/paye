package verification

import (
 "fmt"
 "github.com/alumasinde/budget254-paye-api/internal/payroll/engine"
 "github.com/shopspring/decimal"
)

// Reconcile independently verifies the financial invariant exposed to clients.
func Reconcile(r engine.Result) error {
 gross, e := decimal.NewFromString(r.Gross); if e != nil { return fmt.Errorf("gross: %w", e) }
 paye, e := decimal.NewFromString(r.PAYE); if e != nil { return fmt.Errorf("paye: %w", e) }
 custom, e := decimal.NewFromString(r.CustomDeductions); if e != nil { return fmt.Errorf("custom deductions: %w", e) }
 statutory := decimal.Zero
 for _, c := range r.StatutoryDeductions { v,e:=decimal.NewFromString(c.Amount); if e!=nil{return fmt.Errorf("component %s: %w",c.Code,e)}; statutory=statutory.Add(v) }
 total, e := decimal.NewFromString(r.TotalDeductions); if e != nil { return e }
 net, e := decimal.NewFromString(r.NetSalary); if e != nil { return e }
 expectedTotal := paye.Add(custom).Add(statutory).Round(2)
 if !total.Equal(expectedTotal) { return fmt.Errorf("total deductions mismatch: got %s want %s", total, expectedTotal) }
 expectedNet := gross.Sub(total).Round(2)
 if !net.Equal(expectedNet) { return fmt.Errorf("net salary mismatch: got %s want %s", net, expectedNet) }
 return nil
}
