package token

import (
	"os"
	"path/filepath"
	"runtime"

	log "github.com/sirupsen/logrus"
)

var versions = []string{"discord", "discordcanary", "discordptb"}

func GetToken() (string, error) {
	log.Warnf("discord must not be running to retrieve your token under %v", runtime.GOOS)

	home, def := os.LookupEnv("HOME")
	if !def {
		return "", ErrorNoHomePath
	}

	for _, ver := range versions {
		path := filepath.Join(home, "Library/Application Support", ver, "Local Storage/leveldb")
		log.Debugf("searching for leveldb database in %v", path)

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
