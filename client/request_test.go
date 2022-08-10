package client

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMarshalRequestArgs(t *testing.T) {
	args := RequestArgs{
		IncludeNSFW: true,
		AuthorID:    "12345",
		Limit:       25,
	}
	assert.Equal(t, "?include_nsfw=true&author_id=12345&limit=25", args.MarshalText())
}
