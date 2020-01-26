//+build !windows

package token

import (
	"github.com/pkg/errors"
)

func GetToken() (string, error) {
	return "", errors.New("Token retrieval only currently supported for Windows")
}
