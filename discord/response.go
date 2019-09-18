package discord

import (
	"fmt"
)

func (c Client) Me() (*Me, error) {
	endpoint := endpoints["me"]
	var me Me
	err := c.request("GET", endpoint, nil, &me)
	if err != nil {
		return nil, err
	}

	return &me, nil
}

func (c Client) Channels() ([]Channel, error) {
	endpoint := endpoints["channels"]
	var channels []Channel
	err := c.request("GET", endpoint, nil, &channels)
	if err != nil {
		return nil, err
	}

	return channels, nil
}

func (c Client) ChannelMessages(channel Channel, me *Me, seek int) (*MessageContext, error) {
	endpoint := fmt.Sprintf(endpoints["channel_msgs"], channel.ID, me.ID, seek, messageLimit)
	var results MessageContext
	err := c.request("GET", endpoint, nil, &results)
	if err != nil {
		return nil, err
	}

	return &results, nil
}

func (c Client) Relationships() ([]Relationship, error) {
	endpoint := endpoints["relationships"]
	var relations []Relationship
	err := c.request("GET", endpoint, nil, &relations)
	if err != nil {
		return nil, err
	}

	return relations, nil
}
