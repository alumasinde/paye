package model

import "time"

type Definition struct {
	ID          uint64  `json:"id"`
	Code        string  `json:"code"`
	Name        string  `json:"name"`
	Category    string  `json:"category"`
	Description *string `json:"description,omitempty"`
}

type Parameter struct {
	Name    string  `json:"name"`
	Type    string  `json:"type"`
	Decimal *string `json:"decimal,omitempty"`
	Integer *int64  `json:"integer,omitempty"`
	Boolean *bool   `json:"boolean,omitempty"`
	Text    *string `json:"text,omitempty"`
}

type Band struct {
	From        string  `json:"from"`
	To          *string `json:"to,omitempty"`
	Rate        *string `json:"rate,omitempty"`
	FixedAmount *string `json:"fixed_amount,omitempty"`
	Order       int     `json:"order"`
	Label       *string `json:"label,omitempty"`
}

type Dependency struct {
	RuleCode string `json:"rule_code"`
	Type     string `json:"type"`
}

type Source struct {
	Authority       string     `json:"authority"`
	Title           string     `json:"title"`
	SourceURL       *string    `json:"source_url,omitempty"`
	ReferenceNumber *string    `json:"reference_number,omitempty"`
	PublishedOn     *time.Time `json:"published_on,omitempty"`
}

type ResolvedRule struct {
	Definition
	VersionID            uint64       `json:"version_id"`
	VersionCode          string       `json:"version_code"`
	VersionName          string       `json:"version_name"`
	CalculationMethod    string       `json:"calculation_method"`
	CalculationOrder     int          `json:"calculation_order"`
	AffectsTaxableIncome bool         `json:"affects_taxable_income"`
	AffectsNetPay        bool         `json:"affects_net_pay"`
	EffectiveFrom        time.Time    `json:"effective_from"`
	EffectiveTo          *time.Time   `json:"effective_to,omitempty"`
	Parameters           []Parameter  `json:"parameters"`
	Bands                []Band       `json:"bands"`
	Dependencies         []Dependency `json:"dependencies"`
	Sources              []Source     `json:"sources"`
}
