package validation
import ("testing";"github.com/alumasinde/budget254-paye-api/internal/payrollrules/model")
func TestValidateRejectsDuplicateComponents(t *testing.T){
 x:=model.RuleSet{Code:"KE",Name:"Kenya",EffectiveFrom:"2026-01-01",Components:[]model.Component{
  {ComponentCode:"PAYE",ComponentType:"PAYE_BANDS",FormulaType:"BANDS",Payload:map[string]any{"bands":[]any{map[string]any{"lower":float64(0),"rate":float64(.1)}}}},
  {ComponentCode:"PAYE",ComponentType:"RELIEF",FormulaType:"FIXED",Payload:map[string]any{"amount":float64(1000)}},
 }}
 r:=ValidateRuleSet(x);if r.Valid{t.Fatal("expected invalid result")}
}
