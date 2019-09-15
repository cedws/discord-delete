package discord

import (
	"discord-delete/log"
	"encoding/json"
	"fmt"
	"net/http"
)

const api string = "https://discordapp.com/api/v6"

var endpoints map[string]string = map[string]string{
	"me":            "/users/@me",
	"relationships": "/users/@me/relationships",
	"guilds":        "/users/@me/guilds",
	"guild_msgs": "/guilds/{}/messages/search" +
		"?author_id={}" +
		"&include_nsfw=true" +
		"&offset={}" +
		"&limit={}",
	"channels": "/users/@me/channels",
	"channel_msgs": "/channels/%v/messages/search" +
		"?author_id=%v" +
		"&include_nsfw=true" +
		"&offset=%v" +
		"&limit=%v",
	"delete_msg": "/channels/{}/messages/{}",
}

type Client struct {
	Token string
}

type Me struct {
	Id string `json:"id"`
}

type Channel struct {
	Id string `json:"id"`
}

type Message struct {
	Id        string `json:"id"`
	Hit       bool   `json:"hit"`
	ChannelId string `json:"channel_id"`
	Type      int    `json:"type"`
}

type MessageResults struct {
	TotalResults int         `json:"total_results"`
	Messages     [][]Message `json:"messages"`
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
	for _, elem := range channels {
		results, err := c.ChannelMessages(elem, me)
		if err != nil {
			return err
		}
		for _, ctx := range results.Messages {
			for _, msg := range ctx {
				if !msg.Hit {
					continue
				}
				if msg.Type != 0 {
					continue
				}
				// TODO: Delete the message.
				log.Logger.Printf(msg.Id)
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
	defer res.Body.Close()
	json.NewDecoder(res.Body).Decode(data)

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
	endpoint := fmt.Sprintf(endpoints["channel_msgs"], channel.Id, me.Id, 0, 25)
	results := new(MessageResults)
	err := c.request("GET", endpoint, &results)
	if err != nil {
		return nil, err
	}

	return results, nil
}
