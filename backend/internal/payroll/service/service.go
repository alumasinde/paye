package service

import (
	"context"
	"fmt"
	"time"

	"github.com/shopspring/decimal"

	"github.com/alumasinde/budget254-paye-api/internal/payroll/engine"
	"github.com/alumasinde/budget254-paye-api/internal/payroll/model"
	rulesmodel "github.com/alumasinde/budget254-paye-api/internal/rules/model"
)

// RuleProvider resolves the fully-loaded (parameters + bands) published
// rules in effect on a given date. Repository.ResolvedApplicable satisfies
// this.
type RuleProvider interface {
	ResolvedApplicable(ctx context.Context, date time.Time) ([]rulesmodel.ResolvedRule, error)
}

type Service struct{ Rules RuleProvider }

type Input struct {
	Gross  decimal.Decimal
	Date   time.Time
	Custom []model.CustomDeduction
}

// Calculate resolves the applicable rules for in.Date and runs the generic
// payroll engine (internal/payroll/engine.Calculate) against them,
// converting the result back into model.Result for the HTTP handler/DTO
// layer, which predates the generic engine and expects this shape.
func (s Service) Calculate(ctx context.Context, in Input) (model.Result, error) {
	if !in.Gross.IsPositive() {
		return model.Result{}, fmt.Errorf("gross salary must be greater than zero")
	}
	rules, err := s.Rules.ResolvedApplicable(ctx, in.Date)
	if err != nil {
		return model.Result{}, err
	}
	rules = withoutAllowableDeductionCeilings(rules)

	customTaxable, customNet := decimal.Zero, decimal.Zero
	for _, d := range in.Custom {
		if d.Type == model.TaxableIncome {
			customTaxable = customTaxable.Add(d.Amount)
		} else {
			customNet = customNet.Add(d.Amount)
		}
	}

	out, err := engine.Calculate(engine.Input{Gross: in.Gross, CustomTaxable: customTaxable, CustomNet: customNet}, rules)
	if err != nil {
		return model.Result{}, err
	}

	res := model.Result{Date: in.Date, Gross: in.Gross, RuleVersions: map[string]string{}}
	if v, e := decimal.NewFromString(out.TaxableIncome); e == nil {
		res.TaxableIncome = v
	}
	if v, e := decimal.NewFromString(out.PAYEBeforeRelief); e == nil {
		res.PAYEBeforeRelief = v
	}
	if v, e := decimal.NewFromString(out.Relief); e == nil {
		res.Relief = v
	}
	if v, e := decimal.NewFromString(out.PAYE); e == nil {
		res.PAYE = v
	}
	if v, e := decimal.NewFromString(out.TotalDeductions); e == nil {
		res.TotalDeductions = v
	}
	if v, e := decimal.NewFromString(out.NetSalary); e == nil {
		res.Net = v
	}

	affectsTaxable := map[string]bool{}
	for _, r := range rules {
		res.RuleVersions[r.Code] = r.VersionCode
		affectsTaxable[r.Code] = r.AffectsTaxableIncome
	}
	for _, x := range out.StatutoryDeductions {
		amt, _ := decimal.NewFromString(x.Amount)
		res.Statutory = append(res.Statutory, model.Item{Code: x.Code, Name: x.Name, Amount: amt, ReducesTaxableIncome: affectsTaxable[x.Code]})
	}
	for _, d := range in.Custom {
		res.Custom = append(res.Custom, model.Item{Code: "CUSTOM", Name: d.Name, Amount: d.Amount, ReducesTaxableIncome: d.Type == model.TaxableIncome})
	}
	return res, nil
}

// withoutAllowableDeductionCeilings drops rules in the ALLOWABLE_DEDUCTION
// category (qualifying pension, qualifying mortgage interest, the
// post-retirement medical fund). These are ceilings on a contribution the
// employee actually declares - their CAPPED_PERCENTAGE parameters carry a
// maximum_amount and a text calculation_base ("ACTUAL_QUALIFYING_..."), not
// a rate, because there is no automatic percentage-of-gross to compute. An
// employee only benefits from them by declaring the real amount as a
// TAXABLE_INCOME custom deduction (which the caller is responsible for
// capping at the published maximum - see GET /api/v1/payroll/rules). Feeding
// them into the engine as an ordinary rule fails since it has no rate.
func withoutAllowableDeductionCeilings(rules []rulesmodel.ResolvedRule) []rulesmodel.ResolvedRule {
	out := make([]rulesmodel.ResolvedRule, 0, len(rules))
	for _, r := range rules {
		if r.Category == "ALLOWABLE_DEDUCTION" {
			continue
		}
		out = append(out, r)
	}
	return out
}
