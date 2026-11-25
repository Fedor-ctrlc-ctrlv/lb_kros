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

	baseline, err := audit.DecodeSnapshot(input.Baseline)
	if err != nil {
		return run, conflict("invalid stored baseline: " + err.Error())
	}

	current, err := audit.DecodeSnapshot(input.Current)
	if err != nil {
		return run, conflict("invalid stored snapshot: " + err.Error())
	}

	diff, err := audit.Compare(baseline, current)
	if err != nil {
		return run, conflict(err.Error())
	}

	checks, evaluation, err := evaluatePolicy(
		current,
		diff,
		input.PolicyID,
		input.Policy,
	)
	if err != nil {
		return run, err
	}

	if err := ctx.Err(); err != nil {
		return run, err
	}

	result := Result{
		Diff:   diff,
		Checks: checks,
	}

	summary := Summary{
		PolicyEvaluation: evaluation,
		Changes:          len(diff.Changes),
		BySection:        map[string]int{},
		ByKind:           map[string]int{},
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

	policySnapshot := input.Policy
	if input.PolicyID == nil {
		policySnapshot = nil
	}

	run = Run{
		CurrentID:     id,
		BaselineID:    *input.BaselineID,
		PolicyID:      input.PolicyID,
		Policy:        policySnapshot,
		EngineVersion: engineVersion(),
		Summary:       summary,
	}

	return s.repo.Save(ctx, run, result)
}

func evaluatePolicy(
	current audit.Snapshot,
	diff audit.DiffReport,
	policyID *int64,
	raw json.RawMessage,
) ([]Check, PolicyEvaluation, error) {
	checks := []Check{}

	if policyID == nil {
		return checks, PolicyEvaluation{
			Status: "not_assigned",
			Reason: "No policy is assigned; only snapshot comparison was performed",
		}, nil
	}

	if len(raw) == 0 || string(raw) == "null" {
		return nil, PolicyEvaluation{}, conflict(
			"assigned policy has no engine definition",
		)
	}

	policy, err := audit.DecodePolicy(raw)
	if err != nil {
		return nil, PolicyEvaluation{}, conflict(
			"invalid stored policy: " + err.Error(),
		)
	}

	if len(policy.StateRules) == 0 && len(policy.ChangeRules) == 0 {
		return checks, PolicyEvaluation{
			Status: "no_rules",
			Reason: "The assigned policy contains no rules",
		}, nil
	}

	stateChecks, err := audit.Check(current, policy)
	if err != nil {
		return nil, PolicyEvaluation{}, conflict(err.Error())
	}

	changeChecks, err := audit.CheckChanges(diff, policy)
	if err != nil {
		return nil, PolicyEvaluation{}, conflict(err.Error())
	}

	for _, finding := range stateChecks {
		checks = append(checks, Check{
			Phase:   "state",
			Finding: finding,
		})
	}

	for _, finding := range changeChecks {
		checks = append(checks, Check{
			Phase:   "change",
			Finding: finding,
		})
	}

	return checks, PolicyEvaluation{
		Status: "evaluated",
	}, nil
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
