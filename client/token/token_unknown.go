//go:build !windows && !linux && !darwin

package token

func GetToken() (string, error) {
	return "", ErrorTokenPlatform
}
