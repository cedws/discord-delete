package discord

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"net/http"
	"time"
)

const api string = "https://discordapp.com/api/v6"
const messageLimit int = 25

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
		return errors.Wrap(err, "Error fetching profile information")
	}

	channels, err := c.Channels()
	if err != nil {
		return errors.Wrap(err, "Error fetching channels")
	}
	for _, channel := range channels {
		c.DeleteMessages(me, channel, c.ChannelMessages)
	}

	relationships, err := c.Relationships()
	if err != nil {
		return errors.Wrap(err, "Error fetching relationships")
	}

Relationships:
	for _, relation := range relationships {
		for _, channel := range channels {
			if len(channel.Recipients) != 1 {
				continue Relationships
			}
			// If the relation is the sole recipient in one of the channels we found
			// earlier, skip it.
			if relation.ID == channel.Recipients[0].ID {
				log.Debugf("Skipping resolving relation because the user already has the channel open")
				continue Relationships
			}
		}

		channel, err := c.ChannelRelationship(relation)
		if err != nil {
			return errors.Wrap(err, "Error resolving relationship to channel")
		}

		log.Debugf("Resolved relationship %v to channel %v", relation.ID, channel.ID)

		c.DeleteMessages(me, *channel, c.ChannelMessages)
	}

	guilds, err := c.Guilds()
	if err != nil {
		return errors.Wrap(err, "Error fetching guilds")
	}
	for _, channel := range guilds {
		c.DeleteMessages(me, channel, c.GuildMessages)
	}

	return nil
}

func (c Client) DeleteMessages(me *Me, channel Channel, messages func(channel Channel, me *Me, seek int) (*Messages, error)) error {
	seek := 0

ChannelMessages:
	for {
		results, err := messages(channel, me, seek)
		if err != nil {
			return errors.Wrap(err, "Error fetching messages for channel")
		}
		if len(results.ContextMessages) == 0 {
			log.Infof("No more messages to delete for channel %v", channel.ID)
			break
		}

		// TODO: See if we can remove this label.
	ContextMessages:
		for _, ctx := range results.ContextMessages {
			for _, msg := range ctx {
				if msg.Hit {
					// Only messages of type zero can be deleted.
					// An example of a non-zero type message is a call request.
					if msg.Type == 0 {
						log.Infof("Deleting message %v from channel %v", msg.ID, channel.ID)
						err := c.DeleteMessage(msg)
						if err != nil {
							return err
						}

						// TODO: Try to remove this. We immediately re-retrieve the message list
						// after a deletion as a workaround for a bug wherein some messages would
						// not be deleted. This results in many more requests being made to the server.
						continue ChannelMessages
					} else {
						log.Debugf("Found message of non-zero type, incrementing seek index")
						seek++

						// We've found the message for this context, we can move on to
						// the next.
						continue ContextMessages
					}
				}

				// This is a context message which may or may not be authored
				// by the current user.
				log.Debugf("Skipping context message")
			}

			// We finished iterating the context but didn't find a message
			// authored by the current user.
			return errors.New("No hit message present in message context")
		}
	}

	return nil
}

func (c Client) request(method string, endpoint string, reqData interface{}, resData interface{}) error {
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
	req.Header.Set("Authorization", c.Token)
	req.Header.Set("Content-Type", "application/json")

	res, err := c.HTTPClient.Do(req)
	if err != nil {
		return errors.Wrap(err, "Error sending request")
	}

	defer res.Body.Close()

	switch status := res.StatusCode; {
	case status >= http.StatusInternalServerError:
		return errors.New("Server returned status Internal Server Error")
	case status == http.StatusTooManyRequests:
		data := new(TooManyRequests)
		err := json.NewDecoder(res.Body).Decode(data)
		if err != nil {
			return errors.Wrap(err, "Error decoding response")
		}
		log.Debugf("Server asked us to sleep for %v milliseconds", data.RetryAfter)
		time.Sleep(time.Duration(data.RetryAfter) * time.Millisecond)
		// Try again once we've waited for the period that the server has asked us to.
		return c.request(method, endpoint, reqData, resData)
	case status == http.StatusForbidden:
		log.Debug("Server returned status Forbidden")
	case status == http.StatusUnauthorized:
		return errors.New("Server sent Unauthorized, is your token correct?")
	case status == http.StatusBadRequest:
		return errors.New("Server returned status Bad Request")
	case status == http.StatusNoContent:
		log.Debug("Server returned status No Content")
	case status == http.StatusAccepted:
		log.Debug("Server returned status Accepted")
	case status == http.StatusOK:
		err := json.NewDecoder(res.Body).Decode(resData)
		if err != nil {
			return errors.Wrap(err, "Error decoding response")
		}
	default:
		return errors.New(fmt.Sprintf("Server sent unhandled status code %v", res.StatusCode))
	}

	return nil
}

type Me struct {
	ID string `json:"id"`
}

type Channel struct {
	ID         string         `json:"id"`
	Recipients []Relationship `json:"recipients"`
	Name       string         `json:"name"`
}

type Relationship struct {
	ID string `json:"id"`
}

type Message struct {
	ID        string `json:"id"`
	Hit       bool   `json:"hit"`
	ChannelID string `json:"channel_id"`
	Type      int    `json:"type"`
}

type Messages struct {
	TotalResults    int         `json:"total_results"`
	ContextMessages [][]Message `json:"messages"`
}

type TooManyRequests struct {
	RetryAfter int `json:"retry_after"`
}
