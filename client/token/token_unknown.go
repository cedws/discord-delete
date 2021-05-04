//+build !windows,!linux

package token

import (
	"github.com/pkg/errors"
)

func GetToken() (string, error) {
	return "", ErrorTokenPlatform
}
