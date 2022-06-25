package client

import (
	"errors"
	"time"

	"github.com/adversarialtools/discord-delete/client/snowflake"

	log "github.com/sirupsen/logrus"
)

var ErrorInvalidDuration = errors.New("error parsing duration")

const day = time.Hour * 24

func (c *Client) SetDryRun(dryRun bool) {
	c.dryRun = dryRun
}

func (c *Client) SetSkipChannels(skipChannels []string) {
	c.skipChannels = skipChannels
}

func (c *Client) SetSkipPinned(skipPinned bool) {
	c.skipPinned = skipPinned
}

func (c *Client) SetMinAge(minAge uint) error {
	t := time.Now().Add(-time.Duration(minAge) * day)
	millis := t.UnixNano() / int64(time.Millisecond)

	c.maxID = snowflake.ToSnowflake(millis)
	log.Debugf("message maximum ID must be %v", c.maxID)

	return nil
}

func (c *Client) SetMaxAge(maxAge uint) error {
	t := time.Now().Add(-time.Duration(maxAge) * day)
	millis := t.UnixNano() / int64(time.Millisecond)

	c.minID = snowflake.ToSnowflake(millis)
	log.Debugf("message minimum ID must be %v", c.minID)

	return nil
}
