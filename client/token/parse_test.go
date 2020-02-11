package token

import (
	"testing"
)

func TestUnquotedBadToken(t *testing.T) {
	_, err := parseToken(`DEADBEEF`)
	if err == nil {
		t.Fatal("Token was invalid but error was nil")
	}
}

func TestPrefixedGoodToken(t *testing.T) {
	_, err := parseToken(`DEAD"BEEF"`)
	if err != nil {
		t.Fatal("Token was valid but error was returned")
	}
}

func TestParsedToken(t *testing.T) {
	tok, _ := parseToken(`DEAD"mfa.BEEF"`)
	if tok != "mfa.BEEF" {
		t.Fatal("Token was not parsed correctly")
	}
}
