package model

import "time"

type RuleSet struct {
    PublicID string `json:"id"`
    Code string `json:"code"`
    Name string `json:"name"`
    Jurisdiction string `json:"jurisdiction"`
    EffectiveFrom string `json:"effective_from"`
    EffectiveTo *string `json:"effective_to,omitempty"`
    Status string `json:"status"`
    VersionNumber int `json:"version_number"`
    SourceReference string `json:"source_reference,omitempty"`
    SourceNotes string `json:"source_notes,omitempty"`
    Components []Component `json:"components,omitempty"`
    CreatedAt time.Time `json:"created_at"`
    UpdatedAt time.Time `json:"updated_at"`
}

type Component struct {
    PublicID string `json:"id"`
    ComponentCode string `json:"component_code"`
    ComponentType string `json:"component_type"`
    Name string `json:"name"`
    CalculationOrder int `json:"calculation_order"`
    ReducesTaxableIncome bool `json:"reduces_taxable_income"`
    ReducesNetPay bool `json:"reduces_net_pay"`
    FormulaType string `json:"formula_type"`
    Payload map[string]any `json:"payload"`
    IsActive bool `json:"is_active"`
}
