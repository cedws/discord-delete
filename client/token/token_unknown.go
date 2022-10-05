//go:build !darwin

package token

func GetToken() (string, error) {
	return "", ErrorTokenPlatform
}
