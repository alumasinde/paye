package engine
import ("fmt"; "sort"; "github.com/shopspring/decimal"; "github.com/alumasinde/budget254-paye-api/internal/payroll/model")
type Engine struct{}
func (Engine) Progressive(in decimal.Decimal,bands []model.Band)(decimal.Decimal,[]model.TraceBand,error){
 sort.Slice(bands,func(i,j int)bool{return bands[i].From.LessThan(bands[j].From)})
 tax:=decimal.Zero; trace:=[]model.TraceBand{}
 for _,b:=range bands{
  if in.LessThanOrEqual(b.From){continue}
  upper:=in; if b.To!=nil && upper.GreaterThan(*b.To){upper=*b.To}
  portion:=upper.Sub(b.From); if portion.IsNegative(){continue}
  t:=portion.Mul(b.Rate).Div(decimal.NewFromInt(100)); tax=tax.Add(t)
  var to *decimal.Decimal; if b.To!=nil{v:=*b.To;to=&v}
  trace=append(trace,model.TraceBand{From:b.From,To:to,Rate:b.Rate,Tax:t})
 }
 return tax.Round(2),trace,nil
}
func (Engine) Percentage(base decimal.Decimal,rule model.Rule)(decimal.Decimal,error){
 rate,ok:=rule.Parameters["rate"];if !ok{return decimal.Zero,fmt.Errorf("%s missing rate",rule.Code)}
 v:=base.Mul(rate).Div(decimal.NewFromInt(100))
 if min,ok:=rule.Parameters["minimum"];ok&&v.LessThan(min){v=min}
 if cap,ok:=rule.Parameters["maximum"];ok&&v.GreaterThan(cap){v=cap}
 return v.Round(2),nil
}
