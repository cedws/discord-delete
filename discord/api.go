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
const messageLimit int = 25

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
	Token      string
	HTTPClient http.Client
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

func New(token string) (c Client) {
	return Client{
		Token:      token,
		HTTPClient: http.Client{},
	}
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
		offset := 0

		for {
			results, err := c.ChannelMessages(channel, me, offset)
			if err != nil {
				return err
			}
			if len(results.ContextMessages) == 0 {
				log.Logger.Infof("No more messages to delete for channel %v", channel.ID)
				break
			}
			for _, ctx := range results.ContextMessages {
				for _, msg := range ctx {
					if !msg.Hit {
						continue
					}
					if msg.Type != 0 {
						offset++
						continue
					}
					log.Logger.Infof("Deleting message %v", msg.ID)
					c.DeleteMessage(channel, msg)
				}
			}
		}
	}

	return nil
}

func (c Client) request(method string, endpoint string, data interface{}) error {
	url := api + endpoint

	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", c.Token)

	log.Logger.Debugf("%v %v", method, url)

	res, err := c.HTTPClient.Do(req)
	if err != nil {
		return err
	}

	defer res.Body.Close()

	switch status := res.StatusCode; {
	case status >= 500:
		return errors.New("Server sent Internal Server Error")
	case status == 429:
		var data TooManyRequests
		err := json.NewDecoder(res.Body).Decode(data)
		if err != nil {
			return err
		}
		log.Logger.Debugf("Server asked us to sleep for %v milliseconds", data.RetryAfter)
		time.Sleep(time.Duration(data.RetryAfter) * time.Millisecond)
		// Try again once we've waited for the period that the server has asked us to.
		return c.request(method, endpoint, data)
	case status == 403:
		log.Logger.Info("Server sent Forbidden")
	case status == 401:
		return errors.New("Server sent Unauthorized, is your token correct?")
	case status == 204:
		break
	case status == 200:
		err := json.NewDecoder(res.Body).Decode(data)
		if err != nil {
			return err
		}
	default:
		log.Logger.Debugf("Unhandled status code %v", res.StatusCode)
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

func (c Client) ChannelMessages(channel Channel, me *Me, offset int) (*MessageResults, error) {
	endpoint := fmt.Sprintf(endpoints["channel_msgs"], channel.ID, me.ID, offset, messageLimit)
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
