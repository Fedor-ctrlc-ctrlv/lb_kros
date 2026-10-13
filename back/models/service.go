package models

type Service struct {
	ID      int64 `json:"id"`
	AuditID int64 `json:"audit_id"`

	Name        string `json:"name"`
	Description string `json:"description"`

	State   string `json:"state"`
	Enabled bool   `json:"enabled"`

	ExecPath string `json:"exec_path"`
}

func (s Service) IsRunning() bool {
	return s.State == "active"
}

func (s Service) IsEnabled() bool {
	return s.Enabled
}

func (s Service) ConfigurationChanged(other Service) bool {
	return s.Name != other.Name ||
		s.ExecPath != other.ExecPath ||
		s.Enabled != other.Enabled
}

func (s Service) StateChanged(other Service) bool {
	return s.State != other.State
}
