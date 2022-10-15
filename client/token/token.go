package token

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"

	log "github.com/sirupsen/logrus"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/opt"
)

const (
	v10Prefix   = "v10"
	dpapiPrefix = "DPAPI"
)

var (
	ErrorTokenRetrieve = errors.New("token: error retrieving token from database")
	ErrorTokenPlatform = errors.New("token: retrieval not supported on this platform yet")
	ErrorNoHomePath    = errors.New("token: HOME path not set in environment")
	ErrorNoAppdataPath = errors.New("token: APPDATA path not set in environment")
)

var tokenKeys = []string{
	"_https://discord.com\x00\x01tokens",
	"_https://ptb.discord.com\x00\x01tokens",
	"_https://canary.discord.com\x00\x01tokens",
}

type SafeStorageTokens map[string]string

func getSafeStorageTokens(path string) (SafeStorageTokens, error) {
	db, err := leveldb.OpenFile(path, &opt.Options{
		ReadOnly: true,
	})
	if err != nil {
		return nil, fmt.Errorf("token: error opening database: %w", err)
	}
	defer db.Close()

	for _, key := range tokenKeys {
		log.Debugf("looking for token under key %v", strconv.Quote(key))

		data, err := db.Get([]byte(key), nil)
		if string(data) == "" || err != nil {
			continue
		}

		var tokens SafeStorageTokens
		if err := json.Unmarshal(data[1:], &tokens); err != nil {
			continue
		}

		return tokens, nil
	}

	return nil, ErrorTokenRetrieve
}
