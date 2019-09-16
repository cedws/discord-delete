package discord

import (
	"discord-delete/log"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"
)

const api string = "https://discordapp.com/api/v6"

var endpoints map[string]string = map[string]string{
	"me":            "/users/@me",
	"relationships": "/users/@me/relationships",
	"guilds":        "/users/@me/guilds",
	"guild_msgs": "/guilds/{}/messages/search" +
		"?author_id=%v" +
		"&include_nsfw=true" +
		"&offset=%v" +
		"&limit=%v",
	"channels": "/users/@me/channels",
	"channel_msgs": "/channels/%v/messages/search" +
		"?author_id=%v" +
		"&include_nsfw=true" +
		"&offset=%v" +
		"&limit=%v",
	"delete_msg": "/channels/%v/messages/%v",
}

type Client struct {
	Token string
}

type Me struct {
	ID string `json:"id"`
}

type Channel struct {
	ID string `json:"id"`
}

type Message struct {
	ID        string `json:"id"`
	Hit       bool   `json:"hit"`
	ChannelID string `json:"channel_id"`
	Type      int    `json:"type"`
}

type MessageResults struct {
	TotalResults    int         `json:"total_results"`
	ContextMessages [][]Message `json:"messages"`
}

type TooManyRequests struct {
	RetryAfter int `json:"retry_after"`
}

func (c Client) PartialDelete() error {
	me, err := c.Me()
	if err != nil {
		return err
	}

	channels, err := c.Channels()
	if err != nil {
		return err
	}

	for _, channel := range channels {
		results, err := c.ChannelMessages(channel, me)
		if err != nil {
			return err
		}
		if len(results.ContextMessages) == 0 {
			continue
		}
		for _, ctx := range results.ContextMessages {
			for _, msg := range ctx {
				if !msg.Hit {
					continue
				}
				if msg.Type != 0 {
					continue
				}
				c.DeleteMessage(channel, msg)
				log.Logger.Printf(msg.ID)
			}
		}
	}

	return nil
}

func (c Client) request(method string, endpoint string, data interface{}) error {
	url := fmt.Sprintf("%v%v", api, endpoint)
	// TODO: Reuse Client instead of reinitialising it for each call.
	client := &http.Client{}
	req, _ := http.NewRequest(method, url, nil)
	req.Header.Set("Authorization", c.Token)
	res, err := client.Do(req)
	if err != nil {
		return err
	}

	switch status := res.StatusCode; {
	case status >= 500:
		return errors.New("Server sent status 500")
	case status == 429:
		data := new(TooManyRequests)
		json.NewDecoder(res.Body).Decode(&data)
		if err != nil {
			return err
		}
		time.Sleep(time.Duration(data.RetryAfter) * time.Millisecond)
		c.request(method, endpoint, data)
	case status == 200:
		defer res.Body.Close()
		json.NewDecoder(res.Body).Decode(data)
	}

	return nil
}

func (c Client) Me() (*Me, error) {
	endpoint := endpoints["me"]
	me := new(Me)
	err := c.request("GET", endpoint, &me)
	if err != nil {
		return nil, err
	}

	return me, nil
}

func (c Client) Channels() ([]Channel, error) {
	endpoint := endpoints["channels"]
	var channels []Channel
	err := c.request("GET", endpoint, &channels)
	if err != nil {
		return nil, err
	}

	return channels, nil
}

func (c Client) ChannelMessages(channel Channel, me *Me) (*MessageResults, error) {
	endpoint := fmt.Sprintf(endpoints["channel_msgs"], channel.ID, me.ID, 0, 25)
	results := new(MessageResults)
	err := c.request("GET", endpoint, &results)
	if err != nil {
		return nil, err
	}

	return results, nil
}

func (c Client) DeleteMessage(channel Channel, msg Message) error {
	endpoint := fmt.Sprintf(endpoints["delete_msg"], channel.ID, msg.ID)
	err := c.request("DELETE", endpoint, nil)
	if err != nil {
		return err
	}

	return nil
}
