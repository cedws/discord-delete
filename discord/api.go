package discord

import (
	"bytes"
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

	relationships, err := c.Relationships()
	if err != nil {
		return err
	}

Relationships:
	for _, relation := range relationships {
		channel, err := c.RelationshipChannel(relation)
		if err != nil {
			return err
		}
		log.Logger.Debugf("Resolved relationship %v to channel %v", relation.ID, channel.ID)

		for _, c := range channels {
			if c == *channel {
				continue Relationships
			}
		}

		log.Logger.Debugf("Found hidden relationship channel")
		channels = append(channels, *channel)
	}

	log.Logger.Debugf("Finished searching for channels")

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
						// This is a context message which may or may not be authored
						// by the current user.
						log.Logger.Debugf("Skipping context message")
						continue
					}
					if msg.Type != 0 {
						// Only messages of type 0 can be deleted.
						// An example of a non-zero type message is a call request.
						log.Logger.Debugf("Found message of non-zero type, incrementing offset")
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

func (c Client) request(method string, endpoint string, reqData interface{}, resData interface{}) error {
	url := api + endpoint
	log.Logger.Debugf("%v %v", method, url)

	buffer := new(bytes.Buffer)
	if reqData != nil {
		err := json.NewEncoder(buffer).Encode(reqData)
		if err != nil {
			return err
		}
	}
	req, err := http.NewRequest(method, url, buffer)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", c.Token)
	req.Header.Set("Content-Type", "application/json")

	res, err := c.HTTPClient.Do(req)
	if err != nil {
		return err
	}

	defer res.Body.Close()

	switch status := res.StatusCode; {
	case status >= http.StatusInternalServerError:
		return errors.New("Server sent Internal Server Error")
	case status == http.StatusTooManyRequests:
		var data TooManyRequestsResponse
		err := json.NewDecoder(res.Body).Decode(resData)
		if err != nil {
			return err
		}
		log.Logger.Debugf("Server asked us to sleep for %v milliseconds", data.RetryAfter)
		time.Sleep(time.Duration(data.RetryAfter) * time.Millisecond)
		// Try again once we've waited for the period that the server has asked us to.
		return c.request(method, endpoint, reqData, resData)
	case status == http.StatusForbidden:
		log.Logger.Info("Server sent Forbidden")
	case status == http.StatusUnauthorized:
		return errors.New("Server sent Unauthorized, is your token correct?")
	case status == http.StatusBadRequest:
		return errors.New("Client sent Bad Request")
	case status == http.StatusNoContent:
		break
	case status == http.StatusOK:
		err := json.NewDecoder(res.Body).Decode(resData)
		if err != nil {
			return err
		}
	default:
		return errors.New(fmt.Sprintf("Server sent unhandled status code %v", res.StatusCode))
	}

	return nil
}

func (c Client) Me() (*MeResponse, error) {
	endpoint := endpoints["me"]
	me := new(MeResponse)
	err := c.request("GET", endpoint, nil, &me)
	if err != nil {
		return nil, err
	}

	return me, nil
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

func (c Client) RelationshipChannel(relation RelationshipResponse) (*Channel, error) {
	endpoint := endpoints["channels"]
	channel := new(Channel)
	recipients := ChannelRequest{
		Recipients: []string{relation.ID},
	}
	err := c.request("POST", endpoint, recipients, &channel)
	if err != nil {
		return nil, err
	}

	return channel, nil
}

func (c Client) Relationships() ([]RelationshipResponse, error) {
	endpoint := endpoints["relationships"]
	var relations []RelationshipResponse
	err := c.request("GET", endpoint, nil, &relations)
	if err != nil {
		return nil, err
	}

	return relations, nil
}

func (c Client) ChannelMessages(channel Channel, me *MeResponse, offset int) (*MessageContextResponse, error) {
	endpoint := fmt.Sprintf(endpoints["channel_msgs"], channel.ID, me.ID, offset, messageLimit)
	results := new(MessageContextResponse)
	err := c.request("GET", endpoint, nil, &results)
	if err != nil {
		return nil, err
	}

	return results, nil
}

func (c Client) DeleteMessage(channel Channel, msg Message) error {
	endpoint := fmt.Sprintf(endpoints["delete_msg"], channel.ID, msg.ID)
	err := c.request("DELETE", endpoint, nil, nil)
	if err != nil {
		return err
	}

	return nil
}
