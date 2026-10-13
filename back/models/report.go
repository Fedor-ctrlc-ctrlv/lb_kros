package models

import (
	"fmt"
	"time"
)

type AuditReport struct {
	ID         string      `json:"id"`
	CreatedAt  time.Time   `json:"created_at"`
	Baseline   []FSNode    `json:"baseline"`
	Scanned    []FSNode    `json:"scanned"`
	Violations []Violation `json:"violations"`
}

func (r *AuditReport) CalculateDiff() {
	r.Violations = []Violation{}

	scannedMap := make(map[string]FSNode)
	for _, s := range r.Scanned {
		scannedMap[s.Path] = s
	}

	for _, base := range r.Baseline {
		scanNode, exists := scannedMap[base.Path]
		if !exists {
			v := Violation{Path: base.Path, Type: "deleted"}
			v.EvaluateSeverity()
			r.Violations = append(r.Violations, v)
		} else if base.IsModified(&scanNode) {
			v := Violation{Path: base.Path, Type: "modified"}
			v.EvaluateSeverity()
			r.Violations = append(r.Violations, v)
		}
	}

	for _, s := range r.Scanned {
		if _, exists := scannedMap[s.Path]; !exists {

		}

	}

	r.ID = fmt.Sprintf("report-%d", time.Now().Unix())
	r.CreatedAt = time.Now()
}
