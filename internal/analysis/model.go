package analysis

import (
	"encoding/json"
	"time"

	audit "github.com/Fedor-ctrlc-ctrlv/linux_auditor_engine"
)

type Check struct {
	Phase string `json:"phase"`
	audit.Finding
}

type Summary struct {
	Changes   int            `json:"changes"`
	BySection map[string]int `json:"by_section"`
	ByKind    map[string]int `json:"by_kind"`
	Checks    map[string]int `json:"checks"`
}

type Result struct {
	Diff   audit.DiffReport `json:"diff"`
	Checks []Check          `json:"checks"`
}

type Run struct {
	ID            int64           `json:"id"`
	CurrentID     int64           `json:"current_audit_id"`
	BaselineID    int64           `json:"baseline_audit_id"`
	PolicyID      *int64          `json:"policy_id"`
	Policy        json.RawMessage `json:"policy_snapshot"`
	EngineVersion string          `json:"engine_version"`
	Summary       Summary         `json:"summary"`
	CreatedAt     time.Time       `json:"created_at"`
}

type Inputs struct {
	Current        json.RawMessage
	Baseline       json.RawMessage
	Policy         json.RawMessage
	CurrentID      int64
	BaselineID     *int64
	PolicyID       *int64
	CurrentStatus  string
	BaselineStatus *string
	LegacyEnabled  bool
}

type Filter struct {
	Section string
	Kind    string
	Status  string
	Phase   string
	Limit   int
	Offset  int
}

type Page struct {
	Items  []json.RawMessage `json:"items"`
	Limit  int               `json:"limit"`
	Offset int               `json:"offset"`
}
