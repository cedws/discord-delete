package token

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestUnquotedBadToken(t *testing.T) {
	_, err := parseToken(`DEADBEEF`)
	assert.NotNil(t, err)
}

func TestPrefixedGoodToken(t *testing.T) {
	_, err := parseToken(`DEAD"BEEF"`)
	assert.Nil(t, err)
}

func TestParsedToken(t *testing.T) {
	tok, _ := parseToken(`DEAD"mfa.BEEF"`)
	assert.Equal(t, "mfa.BEEF", tok)
}
