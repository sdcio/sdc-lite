package types

import "github.com/sdcio/data-server/pkg/tree/types"

type ValidationStatsExport struct {
	*types.ValidationStatOverall
	Target   string   `json:"target"`
	Passed   bool     `json:"passed"`
	Errors   []string `json:"errors"`
	Warnings []string `json:"warnings"`
}
