package client

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestParseDurationDay(t *testing.T) {
	d, _ := parseDuration("30d")
	assert.Equal(t, d, 30*day)
}

func TestParseDurationHour(t *testing.T) {
	d, _ := parseDuration("30h")
	assert.Equal(t, d, 30*time.Hour)
}
