//+build !windows,!linux

package token

import (
	"github.com/pkg/errors"
)

func GetToken() (string, error) {
	return "", errors.New("Token retrieval not supported on this platform yet")
}
