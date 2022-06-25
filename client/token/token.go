package token

import (
	"errors"
)

var (
	ErrorTokenRetrieve = errors.New("error retrieving token from database")
	ErrorTokenPlatform = errors.New("token retrieval not supported on this platform yet")
	ErrorTokenInvalid  = errors.New("token doesn't seem valid")
	ErrorNoHomePath    = errors.New("HOME path not available in environment")
	ErrorNoAppdataPath = errors.New("APPDATA path not available in environment")
)
