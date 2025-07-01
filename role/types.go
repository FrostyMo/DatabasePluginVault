package role

import (
	"time"

	"github.com/hashicorp/vault/sdk/database/dbplugin/v5"
)

// RoleEntry represents a dynamic or static role in the system.
type RoleEntry struct {
	DBName           string                  `json:"db_name"`
	Statements       dbplugin.Statements     `json:"statements"`
	DefaultTTL       time.Duration           `json:"default_ttl"`
	MaxTTL           time.Duration           `json:"max_ttl"`
	CredentialType   dbplugin.CredentialType `json:"credential_type"`
	CredentialConfig map[string]interface{}  `json:"credential_config"`
	StaticAccount    *StaticAccount          `json:"static_account" mapstructure:"static_account"`
}

// StaticAccount represents a statically managed credential.
type StaticAccount struct {
	Username                 string        `json:"username"`
	Password                 string        `json:"password"`
	PrivateKey               []byte        `json:"private_key"`
	LastVaultRotation        time.Time     `json:"last_vault_rotation"`
	RotationPeriod           time.Duration `json:"rotation_period"`
	RevokeUserOnDelete       bool          `json:"revoke_user_on_delete"`
	PasswordPolicy           string        `json:"password_policy"`
	AutorotateFailureRetries int           `json:"autorotate_failure_retries"`
	LastAutorotateAttempt    time.Time     `json:"last_autorotate_attempt"`
	LastAutorotateError      string        `json:"last_autorotate_error"`
}

// NextRotationTime returns when the static credential is due for rotation.
func (s *StaticAccount) NextRotationTime() time.Time {
	return s.LastVaultRotation.Add(s.RotationPeriod)
}

// CredentialTTL calculates how long this credential is still valid.
func (s *StaticAccount) CredentialTTL() time.Duration {
	return time.Until(s.NextRotationTime())
}
