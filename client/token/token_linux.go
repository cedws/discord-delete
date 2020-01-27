//+build linux

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
	home, def := os.LookupEnv("HOME")
	if !def {
		return "", errors.New("HOME path wasn't specified in environment")
	}
	path := filepath.Join(home, ".config/discord/Local Storage/leveldb")

	db, err := leveldb.OpenFile(path, &opt.Options{
		ReadOnly: true,
	})
	if err != nil {
		return "", errors.Wrap(err, "Couldn't open database")
	}
	defer db.Close()

	data, err := db.Get([]byte(tokenKey), nil)
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
