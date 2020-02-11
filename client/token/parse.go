package token

import (
	"github.com/pkg/errors"
	"regexp"
)

func parseToken(data string) (string, error) {
	reg := regexp.MustCompile(`"(.*)"`)
	match := reg.FindStringSubmatch(data)
	if len(match) < 1 {
		return "", errors.New("Token doesn't seem valid")
	}
	return match[1], nil
}
