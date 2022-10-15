package token

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/billgraziano/dpapi"
	log "github.com/sirupsen/logrus"
)

type LocalState struct {
	OsCrypt struct {
		EncryptedKey string `json:"encrypted_key"`
	} `json:"os_crypt"`
}

var versions = []string{"Discord", "discordcanary", "discordptb"}

func GetToken() (string, error) {
	log.Warnf("discord must not be running to retrieve your token under %v", runtime.GOOS)

	appdata, def := os.LookupEnv("APPDATA")
	if !def {
		return "", ErrorNoAppdataPath
	}

	for _, ver := range versions {
		path := filepath.Join(appdata, ver, "Local Storage/leveldb")
		log.Debugf("searching for leveldb database in %v", path)

		safeTokens, err := getSafeStorageTokens(path)
		if err != nil {
			// try another database
			log.Error(err)
			continue
		}

		path = filepath.Join(appdata, ver, "Local State")
		key, err := getDecryptionKey(path)
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

func getDecryptionKey(path string) ([]byte, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("token: error reading local state: %w", err)
	}
	defer file.Close()

	var localState LocalState
	if err := json.NewDecoder(file).Decode(&localState); err != nil {
		return nil, fmt.Errorf("token: error unmarshaling local state: %w", err)
	}

	decoded, err := base64.StdEncoding.DecodeString(localState.OsCrypt.EncryptedKey)
	if err != nil {
		return nil, fmt.Errorf("token: error decoding: %w", err)
	}

	decrypted, err := dpapi.DecryptBytes(decoded[len(dpapiPrefix):])
	if err != nil {
		return nil, fmt.Errorf("token: error decrypting with dpapi: %w", err)
	}

	return []byte(decrypted), nil
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

	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonce := safeTokenBytes[len(v10Prefix) : len(v10Prefix)+12]
	ciphertext := safeTokenBytes[len(v10Prefix)+12:]

	plaintext, err := aesgcm.Open(nil, []byte(nonce), []byte(ciphertext), nil)
	if err != nil {
		return "", err
	}

	return string(plaintext), nil
}
