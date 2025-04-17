package security

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
)

// DefaultSecurityService implements SecurityService
type DefaultSecurityService struct {
	signatureSecret   string
	storeNonce        func(nonce string, expiration time.Duration) error
	getNonce          func(nonce string) (bool, error)
	invalidateNonce   func(nonce string) error
	nonceValidityTime time.Duration
}

// NewSecurityService creates a new security service
func NewSecurityService(
	signatureSecret string,
	nonceValidityTime time.Duration,
	storeNonce func(nonce string, expiration time.Duration) error,
	getNonce func(nonce string) (bool, error),
	invalidateNonce func(nonce string) error,
) SecurityService {
	return &DefaultSecurityService{
		signatureSecret:   signatureSecret,
		storeNonce:        storeNonce,
		getNonce:          getNonce,
		invalidateNonce:   invalidateNonce,
		nonceValidityTime: nonceValidityTime,
	}
}

// GenerateNonce creates a new nonce and stores it
func (s *DefaultSecurityService) GenerateNonce() (string, error) {
	// Generate random nonce
	nonce := uuid.New().String()

	// Store nonce in Redis with expiration
	if err := s.storeNonce(nonce, s.nonceValidityTime); err != nil {
		return "", fmt.Errorf("failed to store nonce: %w", err)
	}

	return nonce, nil
}

// ValidateTimestamp checks if the timestamp is within the valid window
func (s *DefaultSecurityService) ValidateTimestamp(timestamp string, validityWindow time.Duration) error {
	ts, err := strconv.ParseInt(timestamp, 10, 64)
	if err != nil {
		return errors.New("invalid timestamp format")
	}

	// Convert milliseconds to time.Time
	requestTime := time.UnixMilli(ts)
	now := time.Now()

	// Check if timestamp is within validity window
	diff := now.Sub(requestTime)
	if diff < -validityWindow || diff > validityWindow {
		return errors.New("timestamp is outside validity window")
	}

	return nil
}

// ValidateSignature verifies that the signature matches the request parameters
func (s *DefaultSecurityService) ValidateSignature(params map[string]string, signature string) error {
	// Sort parameters by key
	var keys []string
	for k := range params {
		if k != "sign" {
			keys = append(keys, k)
		}
	}
	sort.Strings(keys)

	// Build parameter string
	var paramPairs []string
	for _, k := range keys {
		paramPairs = append(paramPairs, fmt.Sprintf("%s=%s", k, params[k]))
	}
	paramString := strings.Join(paramPairs, "&")

	// Calculate HMAC-SHA256
	h := hmac.New(sha256.New, []byte(s.signatureSecret))
	h.Write([]byte(paramString))
	expectedSign := hex.EncodeToString(h.Sum(nil))

	// Compare with provided signature
	if !hmac.Equal([]byte(expectedSign), []byte(signature)) {
		return errors.New("invalid signature")
	}

	return nil
}

// ValidateNonce checks if the nonce is valid and hasn't been used before
func (s *DefaultSecurityService) ValidateNonce(nonce string) error {
	// Check if nonce exists in Redis
	exists, err := s.getNonce(nonce)
	if err != nil {
		return fmt.Errorf("failed to check nonce: %w", err)
	}

	if !exists {
		return errors.New("invalid or expired nonce")
	}

	// Invalidate nonce after use
	if err := s.invalidateNonce(nonce); err != nil {
		return fmt.Errorf("failed to invalidate nonce: %w", err)
	}

	return nil
}

// GenerateSignature creates a signature for the given parameters
func GenerateSignature(params map[string]string, secret string) string {
	// Sort parameters by key
	var keys []string
	for k := range params {
		if k != "sign" {
			keys = append(keys, k)
		}
	}
	sort.Strings(keys)

	// Build parameter string
	var paramPairs []string
	for _, k := range keys {
		paramPairs = append(paramPairs, fmt.Sprintf("%s=%s", k, params[k]))
	}
	paramString := strings.Join(paramPairs, "&")

	// Calculate HMAC-SHA256
	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(paramString))

	return hex.EncodeToString(h.Sum(nil))
}
