package models

type SSHKey struct {
	ID      int64 `json:"id"`
	AuditID int64 `json:"audit_id"`

	Username string `json:"username"`

	KeyType     string `json:"key_type"`
	Fingerprint string `json:"fingerprint"`
	Comment     string `json:"comment"`

	SourcePath string `json:"source_path"`
}

func (k SSHKey) IsRootKey() bool {
	return k.Username == "root"
}

func (k SSHKey) SameKey(other SSHKey) bool {
	return k.Username == other.Username &&
		k.Fingerprint == other.Fingerprint
}

func (k SSHKey) LocationChanged(other SSHKey) bool {
	return k.SourcePath != other.SourcePath
}
