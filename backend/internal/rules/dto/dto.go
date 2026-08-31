package dto

import "time"

type ResolveManyRequest struct {
	Codes []string `json:"codes"`
}
type ResolvedRuleResponse struct {
	Rule any       `json:"rule"`
	AsOf time.Time `json:"as_of"`
}
