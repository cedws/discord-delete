package client

import (
	"fmt"
)

func (c *Client) Me() (*Me, error) {
	endpoint := endpoints["me"]
	var me Me
	err := c.request("GET", endpoint, nil, &me)

	return &me, err
}

func (c *Client) Channels() ([]Channel, error) {
	endpoint := endpoints["channels"]
	var channels []Channel

	err := c.request("GET", endpoint, nil, &channels)
	return channels, err
}

func (c *Client) ChannelMessages(channel *Channel, me *Me, seek *int) (*Messages, error) {
	endpoint := fmt.Sprintf(
		endpoints["channel_msgs"],
		channel.ID,
		me.ID,
		*seek,
		messageLimit,
	)

	if c.minID > 0 {
		endpoint = fmt.Sprintf("%v&min_id=%v", endpoint, c.minID)
	}

	if c.maxID > 0 {
		endpoint = fmt.Sprintf("%v&max_id=%v", endpoint, c.maxID)
	}

	var results Messages
	err := c.request("GET", endpoint, nil, &results)

	return &results, err
}

func (c *Client) ChannelRelationship(relation *Recipient) (*Channel, error) {
	endpoint := endpoints["channels"]
	recipients := struct {
		Recipients []string `json:"recipients"`
	}{
		[]string{relation.ID},
	}
	var channel Channel

	err := c.request("POST", endpoint, recipients, &channel)
	return &channel, err
}

func (c *Client) Relationships() ([]Relationship, error) {
	endpoint := endpoints["relationships"]
	var relations []Relationship

	err := c.request("GET", endpoint, nil, &relations)
	return relations, err
}

func (c *Client) Guilds() ([]Channel, error) {
	endpoint := endpoints["guilds"]
	var channels []Channel

	err := c.request("GET", endpoint, nil, &channels)
	return channels, err
}

func (c *Client) GuildMessages(channel *Channel, me *Me, seek *int) (*Messages, error) {
	endpoint := fmt.Sprintf(
		endpoints["guild_msgs"],
		channel.ID,
		me.ID,
		*seek,
		messageLimit,
	)

	if c.minID > 0 {
		endpoint = fmt.Sprintf("%v&min_id=%v", endpoint, c.minID)
	}

	if c.maxID > 0 {
		endpoint = fmt.Sprintf("%v&max_id=%v", endpoint, c.maxID)
	}

	var results Messages

	err := c.request("GET", endpoint, nil, &results)
	return &results, err
}
