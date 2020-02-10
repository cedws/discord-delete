//+build windows

package token

import (
	"github.com/pkg/errors"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/opt"
	"os"
	"path/filepath"
)

func GetToken() (tok string, err error) {
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
	defer func() {
		err = errors.Wrap(db.Close(), "Error closing database")
	}()

	data, err := db.Get([]byte(tokenKey), nil)
	if err != nil {
		return "", errors.Wrap(err, "Couldn't retrieve token from database")
	}

	tok, err = parseToken(string(data))

	return
}
