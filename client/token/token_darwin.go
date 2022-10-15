package token

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha1"
	"encoding/base64"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/keybase/go-keychain"
	log "github.com/sirupsen/logrus"
	"golang.org/x/crypto/pbkdf2"
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

		safeTokens, err := getSafeStorageTokens(path)
		if err != nil {
			if !errors.Is(err, os.ErrNotExist) {
				log.Error(err)
			}
			continue
		}

		key, err := getDecryptionKey()
		if err != nil {
			log.Error(err)
			continue
		}

		for _, safeToken := range safeTokens {
			// strip rickroll
			safeToken := strings.TrimPrefix(safeToken, "dQw4w9WgXcQ:")
			token, err := decryptToken(key, safeToken)
			if err != nil {
				// try next token
				log.Error(err)
				continue
			}

			return strings.Trim(token, "\n"), nil
		}
	}

	return "", ErrorTokenRetrieve
}

func getDecryptionKey() ([]byte, error) {
	query := keychain.NewItem()
	query.SetSecClass(keychain.SecClassGenericPassword)
	query.SetAccount("discord")
	query.SetReturnData(true)

	results, err := keychain.QueryItem(query)
	if err != nil {
		return nil, fmt.Errorf("token: %w", err)
	}

	if len(results) == 0 {
		return nil, fmt.Errorf("token: unable to retrieve decryption key from keychain")
	}

	key := pbkdf2.Key(results[0].Data, []byte("saltysalt"), 1003, 128/8, sha1.New)
	return key, nil
}

func decryptToken(key []byte, safeToken string) (string, error) {
	safeTokenBytes, err := base64.StdEncoding.DecodeString(safeToken)
	if err != nil {
		return "", fmt.Errorf("token: error decoding safeStorage token: %w", err)
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	iv := bytes.Repeat([]byte{' '}, 16)
	ciphertext := safeTokenBytes[len(v10Prefix):]

	cbc := cipher.NewCBCDecrypter(block, iv)
	cbc.CryptBlocks(ciphertext, ciphertext)

	return string(ciphertext), nil
}
