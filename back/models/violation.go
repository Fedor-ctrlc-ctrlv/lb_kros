package models

import (
	"strings"
)

type Violation struct {
	Path        string `json:"path"`
	Type        string `json:"type"`
	IsCritical  bool   `json:"is_critical"`
	Description string `json:"description"`
}

func (v *Violation) EvaluateSeverity() {
	criticalPaths := []string{"/etc/passwd", "/etc/shadow", "/root/.ssh"}
	for _, cp := range criticalPaths {
		if strings.Contains(v.Path, cp) {
			v.IsCritical = true
			v.Description = "CRITICAL: System security file modified!"
			return
		}
	}
	v.IsCritical = false
	v.Description = "Standard file change detected"
}
