package client

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestToSnowflake(t *testing.T) {
	v := toSnowflake(1619910000000)
	assert.Equal(t, int64(838188033638400000), v)
}

func TestFromSnowflake(t *testing.T) {
	v := fromSnowflake(838188033638400000)
	assert.Equal(t, int64(1619910000000), v)
}
