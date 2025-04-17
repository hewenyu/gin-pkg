package nonce

import "time"

type NonceService interface {
	StoreNonce(nonce string, expiration time.Duration) error
	GetNonce(nonce string) (bool, error)
	Close() error
}
