package client

import (
	"errors"
	log "github.com/sirupsen/logrus"
	"strconv"
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

func (c *Client) SetMinAge(minAge string) error {
	dur, err := parseDuration(minAge)
	if err != nil {
		return err
	}

	t := time.Now().Add(-dur)
	millis := t.UnixNano() / int64(time.Millisecond)

	c.maxID = toSnowflake(millis)
	log.Debugf("Message maximum ID must be %v", c.maxID)

	return nil
}

func (c *Client) SetMaxAge(maxAge string) error {
	dur, err := parseDuration(maxAge)
	if err != nil {
		return err
	}

	t := time.Now().Add(-dur)
	millis := t.UnixNano() / int64(time.Millisecond)

	c.minID = toSnowflake(millis)
	log.Debugf("Message minimum ID must be %v", c.minID)

	return nil
}

func parseDuration(duration string) (time.Duration, error) {
	if len(duration) == 0 {
		return 0, ErrorInvalidDuration
	}

	last := duration[len(duration)-1:]
	mult, err := strconv.Atoi(duration[:len(duration)-1])
	if err != nil {
		return 0, ErrorInvalidDuration
	}

	switch last {
	case "d":
		return time.Duration(mult) * day, nil
	default:
		return time.ParseDuration(duration)
	}

	return 0, nil
}
