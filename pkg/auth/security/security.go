package security

import "time"

// SecurityService defines the interface for security operations
type SecurityService interface {
	GenerateNonce() (string, error)
	ValidateTimestamp(timestamp string, validityWindow time.Duration) error
	ValidateSignature(params map[string]string, signature string) error
	ValidateNonce(nonce string) error
}
