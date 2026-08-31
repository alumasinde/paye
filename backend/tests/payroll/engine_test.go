package payroll_test

import (
 "testing"
 "github.com/alumasinde/budget254-paye-api/internal/payroll/engine"
 "github.com/alumasinde/budget254-paye-api/internal/payroll/verification"
 "github.com/alumasinde/budget254-paye-api/internal/rules/model"
 "github.com/shopspring/decimal"
)
func s(v string)*string{return &v}
func rule(code, method string, order int, taxable, net bool, params []model.Parameter, bands []model.Band) model.ResolvedRule{return model.ResolvedRule{Definition:model.Definition{Code:code,Name:code},VersionCode:code+"_TEST",CalculationMethod:method,CalculationOrder:order,AffectsTaxableIncome:taxable,AffectsNetPay:net,Parameters:params,Bands:bands}}
func d(name,v string) model.Parameter{return model.Parameter{Name:name,Decimal:s(v)}}
func TestProgressiveBoundaries(t *testing.T){
 paye:=rule("PAYE","PROGRESSIVE_BANDS",20,false,false,nil,[]model.Band{{From:"0",To:s("100"),Rate:s("0.10"),Order:1},{From:"100",To:s("200"),Rate:s("0.20"),Order:2},{From:"200",To:nil,Rate:s("0.30"),Order:3}})
 cases:=[]struct{gross,want string}{{"0","0.00"},{"100","10.00"},{"100.01","10.00"},{"200","30.00"},{"200.01","30.00"},{"250","45.00"}}
 for _,tc:=range cases{r,e:=engine.Calculate(engine.Input{Gross:decimal.RequireFromString(tc.gross)},[]model.ResolvedRule{paye});if e!=nil{t.Fatal(e)};if r.PAYEBeforeRelief!=tc.want{t.Errorf("%s got %s want %s",tc.gross,r.PAYEBeforeRelief,tc.want)}}
}
func TestReliefCannotMakePAYENegative(t *testing.T){
 paye:=rule("PAYE","PERCENTAGE",20,false,false,[]model.Parameter{d("rate","0.10")},nil); relief:=rule("PERSONAL_RELIEF","FIXED_AMOUNT",30,false,false,[]model.Parameter{d("amount","50")},nil)
 r,e:=engine.Calculate(engine.Input{Gross:decimal.RequireFromString("100")},[]model.ResolvedRule{paye,relief});if e!=nil{t.Fatal(e)};if r.PAYE!="0.00"{t.Fatalf("got %s",r.PAYE)}
}
func TestTaxableAndNetDeductionsAndReconciliation(t *testing.T){
 nssf:=rule("NSSF","FIXED_AMOUNT",1,true,true,[]model.Parameter{d("amount","10")},nil)
 paye:=rule("PAYE","PERCENTAGE",20,false,false,[]model.Parameter{d("rate","0.10")},nil)
 r,e:=engine.Calculate(engine.Input{Gross:decimal.RequireFromString("100"),CustomTaxable:decimal.RequireFromString("5"),CustomNet:decimal.RequireFromString("7")},[]model.ResolvedRule{nssf,paye});if e!=nil{t.Fatal(e)}
 if r.TaxableIncome!="85.00"{t.Fatalf("taxable %s",r.TaxableIncome)}
 if r.PAYE!="8.50"{t.Fatalf("paye %s",r.PAYE)}
 if e:=verification.Reconcile(r);e!=nil{t.Fatal(e)}
}
func TestCappedPercentage(t *testing.T){
 r1:=rule("X","CAPPED_PERCENTAGE",1,false,true,[]model.Parameter{d("rate","0.10"),d("upper_earnings_limit","100"),d("maximum_contribution","8")},nil)
 r,e:=engine.Calculate(engine.Input{Gross:decimal.RequireFromString("1000")},[]model.ResolvedRule{r1});if e!=nil{t.Fatal(e)};if r.StatutoryDeductions[0].Amount!="8.00"{t.Fatal(r.StatutoryDeductions[0].Amount)}
}
func TestTieredFixed(t *testing.T){
 r1:=rule("X","TIERED_FIXED_AMOUNT",1,false,true,nil,[]model.Band{{From:"0",To:s("100"),FixedAmount:s("1")},{From:"100",To:s("200"),FixedAmount:s("2")},{From:"200",FixedAmount:s("3")}})
 for _,tc:=range []struct{x,w string}{{"50","1.00"},{"150","2.00"},{"250","3.00"}}{r,e:=engine.Calculate(engine.Input{Gross:decimal.RequireFromString(tc.x)},[]model.ResolvedRule{r1});if e!=nil{t.Fatal(e)};if r.StatutoryDeductions[0].Amount!=tc.w{t.Fatalf("%s",r.StatutoryDeductions[0].Amount)}}
}
