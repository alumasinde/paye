package service

import (
    "errors"
    "sort"
    "github.com/shopspring/decimal"
    "github.com/alumasinde/budget254-paye-api/internal/payrollrules/model"
)

type PreviewBand struct{ Lower decimal.Decimal `json:"lower"`; Upper *decimal.Decimal `json:"upper,omitempty"`; Rate decimal.Decimal `json:"rate"` }
type PreviewResult struct{ Gross decimal.Decimal `json:"gross_salary"`; PAYEBeforeRelief decimal.Decimal `json:"paye_before_relief"`; PersonalRelief decimal.Decimal `json:"personal_relief"`; PAYE decimal.Decimal `json:"paye"`; NetBeforeOtherDeductions decimal.Decimal `json:"net_before_other_deductions"`; Trace []map[string]string `json:"trace"` }

func Preview(x model.RuleSet,gross decimal.Decimal)(PreviewResult,error){
    var bands []PreviewBand;relief:=decimal.Zero
    for _,c:=range x.Components {
        if !c.IsActive {continue}
        if c.ComponentType=="PAYE_BANDS" {
            raw,ok:=c.Payload["bands"].([]any);if !ok{return PreviewResult{},errors.New("invalid bands payload")}
            for _,v:=range raw {m:=v.(map[string]any);lower:=decimal.NewFromFloat(m["lower"].(float64));rate:=decimal.NewFromFloat(m["rate"].(float64));b:=PreviewBand{Lower:lower,Rate:rate};if u,ok:=m["upper"].(float64);ok{z:=decimal.NewFromFloat(u);b.Upper=&z};bands=append(bands,b)}
        }
        if c.ComponentType=="RELIEF" && c.FormulaType=="FIXED" {if a,ok:=c.Payload["amount"].(float64);ok{relief=relief.Add(decimal.NewFromFloat(a))}}
    }
    sort.Slice(bands,func(i,j int)bool{return bands[i].Lower.LessThan(bands[j].Lower)})
    tax:=decimal.Zero;trace:=[]map[string]string{}
    for _,b:=range bands {
        if gross.LessThanOrEqual(b.Lower){continue}
        taxable:=gross.Sub(b.Lower)
        if b.Upper!=nil {width:=b.Upper.Sub(b.Lower);if taxable.GreaterThan(width){taxable=width}}
        part:=taxable.Mul(b.Rate);tax=tax.Add(part)
        trace=append(trace,map[string]string{"lower":b.Lower.StringFixed(2),"taxed_amount":taxable.StringFixed(2),"rate":b.Rate.String(),"tax":part.StringFixed(2)})
    }
    final:=tax.Sub(relief);if final.IsNegative(){final=decimal.Zero}
    return PreviewResult{Gross:gross,PAYEBeforeRelief:tax,PersonalRelief:relief,PAYE:final,NetBeforeOtherDeductions:gross.Sub(final),Trace:trace},nil
}
