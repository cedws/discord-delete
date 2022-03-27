package token

import (
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/opt"
	"regexp"
	"strconv"
)

// Discord moved to https://discord.com as of some time around May 2020
// Their webapp uses discord.com whilst their client still uses discordapp.com
// Hence why we need to try and lookup both variants
// We also support grabbing tokens for PTB/Canary, so look these up too
var tokenKeys = []string{
	"_https://discordapp.com\x00\x01token",
	"_https://discord.com\x00\x01token",
	"_https://ptb.discordapp.com\x00\x01token",
	"_https://ptb.discord.com\x00\x01token",
	"_https://canary.discordapp.com\x00\x01token",
	"_https://canary.discord.com\x00\x01token",
}

func parseToken(data string) (string, error) {
	reg := regexp.MustCompile(`"(.*)"`)
	match := reg.FindStringSubmatch(data)
	if len(match) < 1 {
		return "", ErrorTokenInvalid
	}
	return match[1], nil
}

func searchLevelDB(path string) (string, error) {
	db, err := leveldb.OpenFile(path, &opt.Options{
		ReadOnly: true,
	})
	if err != nil {
		return "", errors.Wrap(err, "Couldn't open database")
	}
	defer func() {
		// Drop error, we don't care if this fails
		db.Close()
	}()

	for _, key := range tokenKeys {
		log.Debugf("Looking for token under key %v", strconv.Quote(key))

		data, err := db.Get([]byte(key), nil)
		// Ignore if token is empty
		if string(data) == "" || err != nil {
			continue
		}

		return parseToken(string(data))
	}

	return "", ErrorTokenRetrieve
}
