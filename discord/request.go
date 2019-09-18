package discord

import (
	"fmt"
)

func (c Client) ChannelRelationship(relation Relationship) (*Channel, error) {
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

func (c Client) DeleteMessage(channel Channel, msg Message) error {
	endpoint := fmt.Sprintf(endpoints["delete_msg"], channel.ID, msg.ID)
	err := c.request("DELETE", endpoint, nil, nil)
	if err != nil {
		return err
	}

	return nil
}
