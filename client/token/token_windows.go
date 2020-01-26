//+build windows

package token

import (
	"github.com/pkg/errors"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/opt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
)

func GetToken() (string, error) {
	appdata, def := os.LookupEnv("APPDATA")
	if !def {
		return "", errors.New("APPDATA path wasn't specified in environment")
	}
	path := filepath.Join(appdata, "Discord/Local Storage/leveldb")

	db, err := leveldb.OpenFile(path, &opt.Options{
		ReadOnly: true,
	})
	if err != nil {
		return "", errors.Wrap(err, "Couldn't open database")
	}
	defer db.Close()

	data, err := db.Get([]byte("_https://discordapp.com\x00\x01token"), nil)
	if err != nil {
		return "", errors.Wrap(err, "Couldn't retrieve token from database")
	}

	// TODO: Try to improve this expression or use capture groups.
	reg := regexp.MustCompile("\"(.*?)\"")
	token, err := strconv.Unquote(string(reg.Find(data)))
	if err != nil {
		return "", errors.Wrap(err, "Couldn't parse token")
	}

	return token, nil
}
