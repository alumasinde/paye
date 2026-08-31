package engine

import (
	"fmt"
	"sort"

	"github.com/shopspring/decimal"

	rulesmodel "github.com/alumasinde/budget254-paye-api/internal/rules/model"
)

// Input is the generic engine's calculation input: a gross salary plus any
// custom (non rule-driven) deductions the caller wants applied on top of
// the resolved statutory rules.
type Input struct {
	Gross         decimal.Decimal
	CustomTaxable decimal.Decimal
	CustomNet     decimal.Decimal
}

// StatutoryItem is one statutory (rule-driven) deduction line in a Result.
type StatutoryItem struct {
	Code   string `json:"code"`
	Name   string `json:"name"`
	Amount string `json:"amount"`
}

// Result is the generic engine's calculation output. Monetary fields are
// decimal strings rounded to 2 places.
type Result struct {
	Gross               string          `json:"gross"`
	TaxableIncome       string          `json:"taxable_income"`
	PAYEBeforeRelief    string          `json:"paye_before_relief"`
	Relief              string          `json:"relief"`
	PAYE                string          `json:"paye"`
	StatutoryDeductions []StatutoryItem `json:"statutory_deductions"`
	CustomDeductions    string          `json:"custom_deductions"`
	TotalDeductions     string          `json:"total_deductions"`
	NetSalary           string          `json:"net_salary"`
}

// Calculate runs a set of resolved payroll rules (as returned by the rules
// service) against a gross salary, producing a full breakdown.
//
// Rules are applied in ascending CalculationOrder. A rule coded "PAYE" is
// treated as the income tax rule and computed last, against taxable income.
// A rule coded "PERSONAL_RELIEF" reduces PAYE and can never take it below
// zero. Every other rule is a statutory deduction computed from its
// CalculationMethod (FIXED_AMOUNT, PERCENTAGE, CAPPED_PERCENTAGE,
// TIERED_FIXED_AMOUNT, PROGRESSIVE_BANDS), applied against gross pay.
func Calculate(in Input, rules []rulesmodel.ResolvedRule) (Result, error) {
	sorted := make([]rulesmodel.ResolvedRule, len(rules))
	copy(sorted, rules)
	sort.SliceStable(sorted, func(i, j int) bool { return sorted[i].CalculationOrder < sorted[j].CalculationOrder })

	taxable := in.Gross
	var statutory []StatutoryItem
	statutoryNetTotal := decimal.Zero
	var payeRule, reliefRule *rulesmodel.ResolvedRule

	for i := range sorted {
		r := sorted[i]
		switch r.Code {
		case "PAYE":
			payeRule = &sorted[i]
			continue
		case "PERSONAL_RELIEF":
			reliefRule = &sorted[i]
			continue
		}
		amount, err := computeAmount(r, in.Gross)
		if err != nil {
			return Result{}, err
		}
		if r.AffectsTaxableIncome {
			taxable = taxable.Sub(amount)
		}
		if r.AffectsNetPay {
			statutoryNetTotal = statutoryNetTotal.Add(amount)
		}
		statutory = append(statutory, StatutoryItem{Code: r.Code, Name: r.Name, Amount: amount.Round(2).StringFixed(2)})
	}

	taxable = taxable.Sub(in.CustomTaxable)
	if taxable.IsNegative() {
		taxable = decimal.Zero
	}

	payeBeforeRelief := decimal.Zero
	if payeRule != nil {
		amount, err := computeAmount(*payeRule, taxable)
		if err != nil {
			return Result{}, err
		}
		payeBeforeRelief = amount
	}

	relief := decimal.Zero
	paye := payeBeforeRelief
	if reliefRule != nil {
		amount, err := ruleParam(*reliefRule, "amount")
		if err != nil {
			return Result{}, err
		}
		relief = amount
		paye = paye.Sub(relief)
		if paye.IsNegative() {
			paye = decimal.Zero
		}
	}

	custom := in.CustomTaxable.Add(in.CustomNet)
	total := paye.Add(statutoryNetTotal).Add(custom).Round(2)
	net := in.Gross.Sub(total).Round(2)

	return Result{
		Gross:               in.Gross.Round(2).StringFixed(2),
		TaxableIncome:       taxable.Round(2).StringFixed(2),
		PAYEBeforeRelief:    payeBeforeRelief.Round(2).StringFixed(2),
		Relief:              relief.Round(2).StringFixed(2),
		PAYE:                paye.Round(2).StringFixed(2),
		StatutoryDeductions: statutory,
		CustomDeductions:    custom.Round(2).StringFixed(2),
		TotalDeductions:     total.StringFixed(2),
		NetSalary:           net.StringFixed(2),
	}, nil
}

func ruleParam(r rulesmodel.ResolvedRule, name string) (decimal.Decimal, error) {
	for _, p := range r.Parameters {
		if p.Name != name {
			continue
		}
		if p.Decimal == nil {
			return decimal.Zero, fmt.Errorf("%s: parameter %q has no decimal value", r.Code, name)
		}
		return decimal.NewFromString(*p.Decimal)
	}
	return decimal.Zero, fmt.Errorf("%s: missing parameter %q", r.Code, name)
}

func computeAmount(r rulesmodel.ResolvedRule, base decimal.Decimal) (decimal.Decimal, error) {
	switch r.CalculationMethod {
	case "FIXED_AMOUNT":
		return ruleParam(r, "amount")
	case "PERCENTAGE":
		rate, err := ruleParam(r, "rate")
		if err != nil {
			return decimal.Zero, err
		}
		return base.Mul(rate).Round(2), nil
	case "PERCENTAGE_WITH_MINIMUM":
		rate, err := ruleParam(r, "rate")
		if err != nil {
			return decimal.Zero, err
		}
		min, err := ruleParam(r, "minimum_amount")
		if err != nil {
			return decimal.Zero, err
		}
		amount := base.Mul(rate).Round(2)
		if amount.LessThan(min) {
			amount = min
		}
		return amount, nil
	case "CAPPED_PERCENTAGE":
		return cappedPercentage(r, base)
	case "TIERED_FIXED_AMOUNT":
		return tieredFixed(r.Bands, base)
	case "PROGRESSIVE_BANDS":
		return progressive(r.Bands, base)
	default:
		return decimal.Zero, fmt.Errorf("%s: unsupported calculation method %q", r.Code, r.CalculationMethod)
	}
}

func cappedPercentage(r rulesmodel.ResolvedRule, base decimal.Decimal) (decimal.Decimal, error) {
	rate, err := ruleParam(r, "rate")
	if err != nil {
		return decimal.Zero, err
	}
	uel, err := ruleParam(r, "upper_earnings_limit")
	if err != nil {
		return decimal.Zero, err
	}
	cap, err := ruleParam(r, "maximum_contribution")
	if err != nil {
		return decimal.Zero, err
	}
	capped := base
	if capped.GreaterThan(uel) {
		capped = uel
	}
	amount := capped.Mul(rate).Round(2)
	if amount.GreaterThan(cap) {
		amount = cap
	}
	return amount, nil
}

func tieredFixed(bands []rulesmodel.Band, base decimal.Decimal) (decimal.Decimal, error) {
	for _, b := range bands {
		from, err := decimal.NewFromString(b.From)
		if err != nil {
			return decimal.Zero, fmt.Errorf("tiered band 'from': %w", err)
		}
		if base.LessThan(from) {
			continue
		}
		if b.To != nil {
			to, err := decimal.NewFromString(*b.To)
			if err != nil {
				return decimal.Zero, fmt.Errorf("tiered band 'to': %w", err)
			}
			if base.GreaterThanOrEqual(to) {
				continue
			}
		}
		if b.FixedAmount == nil {
			return decimal.Zero, fmt.Errorf("tiered band matching %s has no fixed_amount", base)
		}
		return decimal.NewFromString(*b.FixedAmount)
	}
	return decimal.Zero, fmt.Errorf("no tiered band matches amount %s", base)
}

func progressive(bands []rulesmodel.Band, base decimal.Decimal) (decimal.Decimal, error) {
	sorted := make([]rulesmodel.Band, len(bands))
	copy(sorted, bands)
	sort.Slice(sorted, func(i, j int) bool {
		fi, _ := decimal.NewFromString(sorted[i].From)
		fj, _ := decimal.NewFromString(sorted[j].From)
		return fi.LessThan(fj)
	})
	tax := decimal.Zero
	for _, b := range sorted {
		from, err := decimal.NewFromString(b.From)
		if err != nil {
			return decimal.Zero, fmt.Errorf("progressive band 'from': %w", err)
		}
		if base.LessThanOrEqual(from) {
			continue
		}
		upper := base
		if b.To != nil {
			to, err := decimal.NewFromString(*b.To)
			if err != nil {
				return decimal.Zero, fmt.Errorf("progressive band 'to': %w", err)
			}
			if upper.GreaterThan(to) {
				upper = to
			}
		}
		portion := upper.Sub(from)
		if portion.IsNegative() {
			continue
		}
		if b.Rate == nil {
			return decimal.Zero, fmt.Errorf("progressive band starting at %s has no rate", b.From)
		}
		rate, err := decimal.NewFromString(*b.Rate)
		if err != nil {
			return decimal.Zero, fmt.Errorf("progressive band 'rate': %w", err)
		}
		tax = tax.Add(portion.Mul(rate))
	}
	return tax.Round(2), nil
}
