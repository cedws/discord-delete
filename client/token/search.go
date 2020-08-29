package token

import (
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/opt"
	"regexp"
)

// Discord moved to https://discord.com as of some time around May 2020
// Their webapp uses discord.com whilst their client still uses discordapp.com
// Hence why we need to try and lookup both variants
// We also support grabbing tokens for PTB/Canary, so look these up too
var tokenKeys = []string{
	"_https://discord.com\x00\x01token",
	"_https://discordapp.com\x00\x01token",
	"_https://ptb.discord.com\x00\x01token",
	"_https://ptb.discordapp.com\x00\x01token",
	"_https://canary.discord.com\x00\x01token",
	"_https://canary.discordapp.com\x00\x01token",
}

func parseToken(data string) (string, error) {
	reg := regexp.MustCompile(`"(.*)"`)
	match := reg.FindStringSubmatch(data)
	if len(match) < 1 {
		return "", errors.New("Token doesn't seem valid")
	}
	return match[1], nil
}

func searchLevelDB(path string) (tok string, err error) {
	db, err := leveldb.OpenFile(path, &opt.Options{
		ReadOnly: true,
	})
	if err != nil {
		err = errors.Wrap(err, "Couldn't open database")
		return
	}
	defer func() {
		// Drop error, we don't care if this fails
		db.Close()
	}()

	for _, key := range tokenKeys {
		log.Debugf("Looking for token under key %v", key)

		data, err := db.Get([]byte(key), nil)
		// Ignore if token is empty
		if string(data) == "" || err != nil {
			continue
		}

		return parseToken(string(data))
	}

	err = errors.New("Failed to retrieve token from database")
	return
}
