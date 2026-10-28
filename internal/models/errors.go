package models

import "errors"

var (
	ErrHostNotFound  = errors.New("host not found")
	ErrAuditNotFound = errors.New("audit not found")

	ErrSnapshotHostMismatch = errors.New(
		"snapshot hostname does not match host",
	)

	ErrBaselineNotSet    = errors.New("baseline is not set")
	ErrAuditNotCompleted = errors.New("audit is not completed")
	ErrAuditHostMismatch = errors.New("audit belongs to another host")

	ErrPolicyNotFound = errors.New("policy not found")
	ErrRuleNotFound   = errors.New("rule not found")
)
