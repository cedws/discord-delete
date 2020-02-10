package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"net/http"
	"time"
)

const api = "https://discordapp.com/api/v6"
const messageLimit = 25

var endpoints map[string]string = map[string]string{
	"me":            "/users/@me",
	"relationships": "/users/@me/relationships",
	"guilds":        "/users/@me/guilds",
	"guild_msgs": "/guilds/%v/messages/search" +
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
	deletedCount int
	requestCount int
	token        string
	httpClient   http.Client
}

func New(token string) (c Client) {
	return Client{
		token:      token,
		httpClient: http.Client{},
	}
}

func (c *Client) PartialDelete() error {
	me, err := c.Me()
	if err != nil {
		return errors.Wrap(err, "Error fetching profile information")
	}

	channels, err := c.Channels()
	if err != nil {
		return errors.Wrap(err, "Error fetching channels")
	}
	for _, channel := range channels {
		err = c.DeleteFromChannel(me, &channel)
		if err != nil {
			return err
		}
	}

	relationships, err := c.Relationships()
	if err != nil {
		return errors.Wrap(err, "Error fetching relationships")
	}

Relationships:
	for _, relation := range relationships {
		for _, channel := range channels {
			// If the relation is the sole recipient in one of the channels we found
			// earlier, skip it.
			if channel.Type == PrivateChannel && channel.Recipients[0].ID == relation.ID {
				log.Debugf("Skipping resolving relation %v because the user already has the channel open", relation.ID)
				continue Relationships
			}
		}

		channel, err := c.ChannelRelationship(&relation.Recipient)
		if err != nil {
			return errors.Wrap(err, "Error resolving relationship to channel")
		}

		log.Infof("Resolved relationship with '%v' to channel %v", relation.Recipient.Username, channel.ID)

		err = c.DeleteFromChannel(me, channel)
		if err != nil {
			return err
		}
	}

	guilds, err := c.Guilds()
	if err != nil {
		return errors.Wrap(err, "Error fetching guilds")
	}
	for _, channel := range guilds {
		err = c.DeleteFromGuild(me, &channel)
		if err != nil {
			return err
		}
	}

	log.Infof("Finished deleting messages: %v deleted in %v total requests", c.deletedCount, c.requestCount)

	return nil
}

func (c *Client) DeleteFromChannel(me *Me, channel *Channel) error {
	seek := 0

	for {
		results, err := c.ChannelMessages(channel, me, &seek)
		if err != nil {
			return errors.Wrap(err, "Error fetching messages for channel")
		}
		if len(results.ContextMessages) == 0 {
			log.Infof("No more messages to delete for channel %v", channel.ID)
			break
		}

		err = c.DeleteMessages(results, &seek)
		if err != nil {
			return err
		}
	}

	return nil
}

func (c *Client) DeleteFromGuild(me *Me, channel *Channel) error {
	seek := 0

	for {
		results, err := c.GuildMessages(channel, me, &seek)
		if err != nil {
			return errors.Wrap(err, "Error fetching messages for guild")
		}
		if len(results.ContextMessages) == 0 {
			log.Infof("No more messages to delete for guild '%v'", channel.Name)
			break
		}

		err = c.DeleteMessages(results, &seek)
		if err != nil {
			return err
		}
	}

	return nil
}

func (c *Client) DeleteMessages(messages *Messages, seek *int) error {
	for _, ctx := range messages.ContextMessages {
		for _, msg := range ctx {
			if msg.Hit {
				// The message might be an action rather than text. Actions aren't deletable.
				// An example of an action is a call request.
				if msg.Type == UserMessage {
					log.Infof("Deleting message %v from channel %v", msg.ID, msg.ChannelID)
					err := c.DeleteMessage(&msg)
					if err != nil {
						return errors.Wrap(err, "Error deleting message")
					}
					c.deletedCount++
				} else {
					log.Debugf("Found message of non-zero type, incrementing seek index")
					(*seek)++
				}

				break
			}

			// This is a context message which may or may not be authored
			// by the current user.
			log.Debugf("Skipping context message")
		}
	}

	return nil
}

func (c *Client) request(method string, endpoint string, reqData interface{}, resData interface{}) error {
	url := api + endpoint
	log.Debugf("%v %v", method, url)

	buffer := new(bytes.Buffer)
	if reqData != nil {
		err := json.NewEncoder(buffer).Encode(reqData)
		if err != nil {
			return errors.Wrap(err, "Error encoding request data")
		}
	}
	req, err := http.NewRequest(method, url, buffer)
	if err != nil {
		return errors.Wrap(err, "Error building request")
	}
	req.Header.Set("Authorization", c.token)
	req.Header.Set("Content-Type", "application/json")

	res, err := c.httpClient.Do(req)
	if err != nil {
		return errors.Wrap(err, "Error sending request")
	}

	c.requestCount++

	defer func() {
		err := res.Body.Close()
		if err != nil {
			log.Fatal(errors.Wrap(err, "Error closing Body"))
		}
	}()

	log.Debugf("Server returned status %v", http.StatusText(res.StatusCode))

	switch status := res.StatusCode; {
	case status >= http.StatusInternalServerError:
		return errors.New(fmt.Sprintf("Bad status code %v", http.StatusText(res.StatusCode)))
	case status == http.StatusAccepted:
		// Server has sent Accepted because the index hasn't been built yet.
		// Wait a while until the server is ready.
		fallthrough
	case status == http.StatusTooManyRequests:
		data := new(ServerWait)
		err := json.NewDecoder(res.Body).Decode(data)
		if err != nil {
			return errors.Wrap(err, "Error decoding response")
		}
		log.Infof("Server asked us to sleep for %v milliseconds", data.RetryAfter)
		time.Sleep(time.Duration(data.RetryAfter) * time.Millisecond)
		// Try again once we've waited for the period that the server has asked us to.
		return c.request(method, endpoint, reqData, resData)
	case status == http.StatusForbidden:
		break
	case status == http.StatusUnauthorized:
		return errors.New(fmt.Sprintf("Bad status code %v, is your token correct?", http.StatusText(res.StatusCode)))
	case status == http.StatusBadRequest:
		return errors.New(fmt.Sprintf("Bad status code %v", http.StatusText(res.StatusCode)))
	case status == http.StatusNoContent:
		break
	case status == http.StatusOK:
		err := json.NewDecoder(res.Body).Decode(resData)
		if err != nil {
			return errors.Wrap(err, "Error decoding response")
		}
	default:
		return errors.New(fmt.Sprintf("Status code %v is unhandled", http.StatusText(res.StatusCode)))
	}

	return nil
}

const (
	UserMessage = 0
)

const (
	PrivateChannel = 1
)

type Me struct {
	ID string `json:"id"`
}

type Channel struct {
	Type       int         `json:"type"`
	ID         string      `json:"id"`
	Recipients []Recipient `json:"recipients"`
	Name       string      `json:"name,omitempty"`
}

type Recipient struct {
	Username string `json:"username"`
	ID       string `json:"id"`
}

type Relationship struct {
	Type      int       `json:"type"`
	ID        string    `json:"id"`
	Recipient Recipient `json:"user"`
}

type Message struct {
	ID        string `json:"id"`
	Hit       bool   `json:"hit,omitempty"`
	ChannelID string `json:"channel_id"`
	Type      int    `json:"type"`
}

type Messages struct {
	TotalResults    int         `json:"total_results"`
	ContextMessages [][]Message `json:"messages"`
}

type ServerWait struct {
	RetryAfter int `json:"retry_after"`
}
