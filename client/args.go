package client

import (
	"errors"
	log "github.com/sirupsen/logrus"
	"time"
)

var (
	ErrorInvalidDuration = errors.New("Failed to parse duration")
)

const day = time.Hour * 24

func (c *Client) SetDryRun(dryRun bool) {
	c.dryRun = dryRun
}

func (c *Client) SetSkipChannels(skipChannels []string) {
	c.skipChannels = skipChannels
}

func (c *Client) SetMinAge(minAge uint) error {
	t := time.Now().Add(-time.Duration(minAge) * day)
	millis := t.UnixNano() / int64(time.Millisecond)

	c.maxID = toSnowflake(millis)
	log.Debugf("Message maximum ID must be %v", c.maxID)

	return nil
}

func (c *Client) SetMaxAge(maxAge uint) error {
	t := time.Now().Add(-time.Duration(maxAge) * day)
	millis := t.UnixNano() / int64(time.Millisecond)

	c.minID = toSnowflake(millis)
	log.Debugf("Message minimum ID must be %v", c.minID)

	return nil
}
