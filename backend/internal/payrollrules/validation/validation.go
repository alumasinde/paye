package validation

import (
    "fmt"
    "sort"
    "strings"
    "github.com/alumasinde/budget254-paye-api/internal/payrollrules/model"
)

type Result struct {
    Valid bool `json:"valid"`
    Errors []string `json:"errors"`
    Warnings []string `json:"warnings"`
}

func ValidateRuleSet(x model.RuleSet) Result {
    r:=Result{Valid:true}
    fail:=func(s string){r.Valid=false;r.Errors=append(r.Errors,s)}
    if strings.TrimSpace(x.Code)=="" { fail("rule set code is required") }
    if strings.TrimSpace(x.Name)=="" { fail("rule set name is required") }
    if x.EffectiveFrom=="" { fail("effective_from is required") }
    if len(x.Components)==0 { fail("at least one rule component is required") }

    seen:=map[string]bool{}
    payeCount:=0
    for _,c:=range x.Components {
        if c.ComponentCode=="" { fail("component_code is required") }
        if seen[c.ComponentCode] { fail(fmt.Sprintf("duplicate component code: %s",c.ComponentCode)) }
        seen[c.ComponentCode]=true
        if c.FormulaType=="" { fail(fmt.Sprintf("%s formula_type is required",c.ComponentCode)) }
        if c.ComponentType=="PAYE_BANDS" {
            payeCount++
            validateBands(c,&r,fail)
        }
        if c.ComponentType=="STATUTORY_DEDUCTION" && c.FormulaType=="PERCENTAGE" {
            if _,ok:=c.Payload["rate"];!ok { fail(fmt.Sprintf("%s percentage component requires rate",c.ComponentCode)) }
        }
    }
    if payeCount>1 { fail("only one PAYE_BANDS component is allowed per rule set") }
    if payeCount==0 { r.Warnings=append(r.Warnings,"rule set contains no PAYE_BANDS component") }
    return r
}

func validateBands(c model.Component,r *Result,fail func(string)) {
    raw,ok:=c.Payload["bands"].([]any)
    if !ok || len(raw)==0 { fail("PAYE_BANDS requires a non-empty bands array"); return }
    lowers:=[]float64{}
    for _,v:=range raw {
        m,ok:=v.(map[string]any);if !ok { fail("invalid PAYE band");continue }
        rate,ok:=number(m["rate"]);if !ok || rate<0 || rate>1 { fail("each PAYE band rate must be between 0 and 1") }
        lower,ok:=number(m["lower"]);if !ok || lower<0 { fail("each PAYE band lower must be >= 0") }
        lowers=append(lowers,lower)
    }
    sorted:=append([]float64(nil),lowers...);sort.Float64s(sorted)
    for i:=range lowers { if lowers[i]!=sorted[i] { r.Warnings=append(r.Warnings,"PAYE bands are not sorted by lower bound");break } }
}
func number(v any)(float64,bool){switch n:=v.(type){case float64:return n,true;case int:return float64(n),true;default:return 0,false}}
