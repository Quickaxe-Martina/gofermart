package client

import (
	"time"
)

// NewAccrualClient create AccrualClient
func NewAccrualClient(URL string, poolSize int, poolTimeout time.Duration) (AccrualClient, error) {
	return NewHTTPAccrualClient(URL, poolSize, poolTimeout), nil
}
