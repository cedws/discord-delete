package client

import (
	"fmt"
)

func (c *Client) DeleteMessage(msg *Message) error {
	endpoint := fmt.Sprintf(endpoints["delete_msg"], msg.ChannelID, msg.ID)
	err := c.request("DELETE", endpoint, nil, nil)
	return err
}
