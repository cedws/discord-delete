package token

import (
	"os"
	"path/filepath"

	log "github.com/sirupsen/logrus"
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
		if err != nil {
			// Try another database
			log.Debug(err)
			continue
		}

		return tok, nil
	}

	return "", ErrorTokenRetrieve
}
