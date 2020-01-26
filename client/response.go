package client

import (
	"fmt"
)

func (c *Client) Me() (*Me, error) {
	endpoint := endpoints["me"]
	var me Me
	err := c.request("GET", endpoint, nil, &me)
	if err != nil {
		return nil, err
	}

	return &me, nil
}

func (c *Client) Channels() ([]Channel, error) {
	endpoint := endpoints["channels"]
	var channels []Channel
	err := c.request("GET", endpoint, nil, &channels)
	if err != nil {
		return nil, err
	}

	return channels, nil
}

func (c *Client) ChannelMessages(channel *Channel, me *Me, seek *int) (*Messages, error) {
	endpoint := fmt.Sprintf(endpoints["channel_msgs"], channel.ID, me.ID, *seek, messageLimit)
	var results Messages
	err := c.request("GET", endpoint, nil, &results)
	if err != nil {
		return nil, err
	}

	return &results, nil
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
	if err != nil {
		return nil, err
	}

	return &channel, nil
}

func (c *Client) Relationships() ([]Relationship, error) {
	endpoint := endpoints["relationships"]
	var relations []Relationship
	err := c.request("GET", endpoint, nil, &relations)
	if err != nil {
		return nil, err
	}

	return relations, nil
}

func (c *Client) Guilds() ([]Channel, error) {
	endpoint := endpoints["guilds"]
	var channels []Channel
	err := c.request("GET", endpoint, nil, &channels)
	if err != nil {
		return nil, err
	}

	return channels, nil
}

func (c *Client) GuildMessages(channel *Channel, me *Me, seek *int) (*Messages, error) {
	endpoint := fmt.Sprintf(endpoints["guild_msgs"], channel.ID, me.ID, *seek, messageLimit)
	var results Messages
	err := c.request("GET", endpoint, nil, &results)
	if err != nil {
		return nil, err
	}

	return &results, nil
}
