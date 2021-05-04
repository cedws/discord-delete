package token

import (
	"errors"
)

var (
	ErrorTokenRetrieve = errors.New("Failed to retrieve token from database")
	ErrorTokenPlatform = errors.New("Token retrieval not supported on this platform yet")
	ErrorTokenInvalid  = errors.New("Token doesn't seem valid")
	ErrorNoHomePath    = errors.New("HOME path wasn't specified in environment")
	ErrorNoAppdataPath = errors.New("APPDATA path wasn't specified in environment")
)
