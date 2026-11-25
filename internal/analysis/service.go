package analysis

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"runtime/debug"

	audit "github.com/Fedor-ctrlc-ctrlv/linux_auditor_engine"
)

var ErrInvalid = errors.New("invalid input")
var ErrConflict = errors.New("analysis prerequisites are not satisfied")

func conflict(message string) error {
	return fmt.Errorf("%w: %s", ErrConflict, message)
}

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) PutPolicy(
	ctx context.Context,
	id int64,
	raw []byte,
) (json.RawMessage, error) {
	policy, err := audit.DecodePolicy(raw)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalid, err)
	}

	normalized, err := json.Marshal(policy)
	if err != nil {
		return nil, err
	}

	return normalized, s.repo.PutPolicy(ctx, id, normalized)
}

func engineVersion() string {
	const modulePath = "github.com/Fedor-ctrlc-ctrlv/linux_auditor_engine"

	info, ok := debug.ReadBuildInfo()
	if ok {
		for _, dependency := range info.Deps {
			if dependency.Path == modulePath {
				if dependency.Replace != nil {
					return modulePath + " local replacement"
				}
				return modulePath + "@" + dependency.Version
			}
		}
	}
	return modulePath + " unversioned"
}

func (s *Service) Create(
	ctx context.Context,
	id int64,
) (Run, error) {
	var run Run

	input, err := s.repo.Inputs(ctx, id)
	if err != nil {
		return run, err
	}

	if input.BaselineID == nil || input.BaselineStatus == nil {
		return run, conflict("host has no baseline")
	}

	if input.CurrentStatus != "COMPLETED" ||
		*input.BaselineStatus != "COMPLETED" {
		return run, conflict("both snapshots must be completed")
	}

	if input.PolicyID == nil || len(input.Policy) == 0 {
		return run, conflict(
			"assign a policy and configure its engine definition",
		)
	}

	baseline, err := audit.DecodeSnapshot(input.Baseline)
	if err != nil {
		return run, conflict("invalid stored baseline: " + err.Error())
	}

	current, err := audit.DecodeSnapshot(input.Current)
	if err != nil {
		return run, conflict("invalid stored snapshot: " + err.Error())
	}

	policy, err := audit.DecodePolicy(input.Policy)
	if err != nil {
		return run, conflict("invalid stored policy: " + err.Error())
	}

	diff, err := audit.Compare(baseline, current)
	if err != nil {
		return run, conflict(err.Error())
	}

	stateChecks, err := audit.Check(current, policy)
	if err != nil {
		return run, conflict(err.Error())
	}

	changeChecks, err := audit.CheckChanges(diff, policy)
	if err != nil {
		return run, conflict(err.Error())
	}

	if err := ctx.Err(); err != nil {
		return run, err
	}

	result := Result{
		Diff:   diff,
		Checks: []Check{},
	}

	for _, finding := range stateChecks {
		result.Checks = append(result.Checks, Check{
			Phase:   "state",
			Finding: finding,
		})
	}
	for _, finding := range changeChecks {
		result.Checks = append(result.Checks, Check{
			Phase:   "change",
			Finding: finding,
		})
	}

	summary := Summary{
		Changes:   len(diff.Changes),
		BySection: map[string]int{},
		ByKind:    map[string]int{},
		Checks: map[string]int{
			"pass":    0,
			"fail":    0,
			"unknown": 0,
		},
	}

	for _, change := range diff.Changes {
		summary.BySection[change.Section]++
		summary.ByKind[change.Kind]++
	}
	for _, check := range result.Checks {
		summary.Checks[check.Status]++
	}

	run = Run{
		CurrentID:     id,
		BaselineID:    *input.BaselineID,
		PolicyID:      input.PolicyID,
		Policy:        input.Policy,
		EngineVersion: engineVersion(),
		Summary:       summary,
	}

	return s.repo.Save(ctx, run, result)
}

func (s *Service) Get(
	ctx context.Context,
	id int64,
) (Run, error) {
	return s.repo.Get(ctx, id)
}

func (s *Service) Coverage(
	ctx context.Context,
	id int64,
) (json.RawMessage, error) {
	return s.repo.Coverage(ctx, id)
}

func (s *Service) Details(
	ctx context.Context,
	id int64,
	checks bool,
	filter Filter,
) (Page, error) {
	return s.repo.Details(ctx, id, checks, filter)
}

func (s *Service) GetPolicy(
	ctx context.Context,
	id int64,
) (json.RawMessage, error) {
	return s.repo.GetPolicy(ctx, id)
}
