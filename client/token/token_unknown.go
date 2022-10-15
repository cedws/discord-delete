//go:build !darwin && !windows

package token

func GetToken() (string, error) {
	return "", ErrorTokenPlatform
}
