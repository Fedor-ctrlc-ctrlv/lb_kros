package models

type UserRole string

const (
	RoleUser  UserRole = "USER"
	RoleAdmin UserRole = "ADMIN"
)

type AuditStatus string

const (
	AuditPending   AuditStatus = "PENDING"
	AuditCompleted AuditStatus = "COMPLETED"
	AuditFailed    AuditStatus = "FAILED"
)

type Severity string

const (
	SeverityInfo     Severity = "INFO"
	SeverityLow      Severity = "LOW"
	SeverityMedium   Severity = "MEDIUM"
	SeverityHigh     Severity = "HIGH"
	SeverityCritical Severity = "CRITICAL"
)

type FsNodeType string

const (
	FsFile      FsNodeType = "FILE"
	FsDirectory FsNodeType = "DIRECTORY"
	FsSymlink   FsNodeType = "SYMLINK"
	FsOther     FsNodeType = "OTHER"
)

type RuleType string

const (
	RuleUserAdded       RuleType = "USER_ADDED"
	RuleUserRemoved     RuleType = "USER_REMOVED"
	RuleUnauthorizedUID RuleType = "UNAUTHORIZED_UID_ZERO"
	RuleSudoAdded       RuleType = "USER_ADDED_TO_SUDO"

	RuleSUIDAdded     RuleType = "SUID_ADDED"
	RuleSGIDAdded     RuleType = "SGID_ADDED"
	RuleCriticalFile  RuleType = "CRITICAL_FILE_CHANGED"
	RuleWorldWritable RuleType = "WORLD_WRITABLE_FILE"

	RulePortAdded        RuleType = "LISTENING_PORT_ADDED"
	RuleUnauthorizedPort RuleType = "UNAUTHORIZED_PORT"

	RuleServiceAdded   RuleType = "SERVICE_ADDED"
	RuleServiceRemoved RuleType = "SERVICE_REMOVED"

	RuleCronAdded     RuleType = "CRON_JOB_ADDED"
	RuleSSHKeyAdded   RuleType = "SSH_KEY_ADDED"
	RuleSSHKeyRemoved RuleType = "SSH_KEY_REMOVED"
)
