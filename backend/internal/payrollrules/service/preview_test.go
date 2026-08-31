package service
import ("testing";"github.com/shopspring/decimal";"github.com/alumasinde/budget254-paye-api/internal/payrollrules/model")
func TestPreviewNeverReturnsNegativePAYE(t *testing.T){
 x:=model.RuleSet{Components:[]model.Component{
  {ComponentCode:"PAYE",ComponentType:"PAYE_BANDS",FormulaType:"BANDS",IsActive:true,Payload:map[string]any{"bands":[]any{map[string]any{"lower":float64(0),"rate":float64(.1)}}}},
  {ComponentCode:"RELIEF",ComponentType:"RELIEF",FormulaType:"FIXED",IsActive:true,Payload:map[string]any{"amount":float64(1000000)}},
 }}
 r,err:=Preview(x,decimal.NewFromInt(10000));if err!=nil{t.Fatal(err)};if !r.PAYE.Equal(decimal.Zero){t.Fatalf("got %s",r.PAYE)}
}
