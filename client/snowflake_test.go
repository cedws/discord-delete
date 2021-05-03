package client

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestToSnowflake(t *testing.T) {
	v := ToSnowflake(1619910000000)
	assert.Equal(t, uint64(838188033638400000), v)
}

func TestFromSnowflake(t *testing.T) {
	v := FromSnowflake(838188033638400000)
	assert.Equal(t, uint64(1619910000000), v)
}
