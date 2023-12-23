package utils

import (
	"github.com/cenkalti/backoff/v4"
	"time"
)

var DefaultBackOff = backoff.NewExponentialBackOff()

func init() {
	DefaultBackOff.MaxElapsedTime = 6 * time.Hour
	DefaultBackOff.MaxInterval = 1 * time.Minute
}

func BackOff(f func() error) error {
	return backoff.Retry(f, DefaultBackOff)
}
