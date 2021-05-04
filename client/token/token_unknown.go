//+build !windows,!linux

package token

func GetToken() (string, error) {
	return "", ErrorTokenPlatform
}
