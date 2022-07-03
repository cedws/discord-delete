package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/cedws/discord-delete/client/spoof"

	log "github.com/sirupsen/logrus"
)

const (
	api          = "https://discord.com/api/v10"
	messageLimit = 25
)

var endpoints = map[string]string{
	"me":             "/users/@me",
	"relationships":  "/users/@me/relationships",
	"guilds":         "/users/@me/guilds",
	"guild_channels": "/guilds/%v/channels",
	"guild_msgs": "/guilds/%v/messages/search" +
		"?include_nsfw=true" +
		"&author_id=%v" +
		"&offset=%v" +
		"&limit=%v",
	"channels": "/users/@me/channels",
	"channel_msgs": "/channels/%v/messages/search" +
		"?include_nsfw=true" +
		"&author_id=%v" +
		"&offset=%v" +
		"&limit=%v",
	"delete_msg": "/channels/%v/messages/%v",
}

type Client struct {
	deletedCount int
	requestCount int
	token        string
	spoof        spoof.Info
	dryRun       bool
	maxID        int64
	minID        int64
	skipChannels []string
	skipPinned   bool
	httpClient   http.Client
}

func New(token string) (c Client) {
	return Client{
		token:      token,
		spoof:      spoof.RandomInfo(),
		httpClient: http.Client{},
	}
}

func (c *Client) PartialDelete() error {
	me, err := c.Me()
	if err != nil {
		return fmt.Errorf("error fetching profile information: %w", err)
	}

	channels, err := c.Channels()
	if err != nil {
		return fmt.Errorf("error fetching channels: %w", err)
	}

	for _, channel := range channels {
		if err = c.DeleteFromChannel(me, &channel); err != nil {
			return err
		}
	}

	relationships, err := c.Relationships()
	if err != nil {
		return fmt.Errorf("error fetching relationships: %w", err)
	}

Relationships:
	for _, relation := range relationships {
		for _, channel := range channels {
			// If the relation is the sole recipient in one of the channels we found
			// earlier, skip it.
			if channel.Type == DirectChannel && channel.Recipients[0].ID == relation.ID {
				log.Debugf("skipping resolving relation %v because the user already has the channel open", relation.ID)
				continue Relationships
			}
		}

		channel, err := c.ChannelRelationship(&relation.Recipient)
		if err != nil {
			return fmt.Errorf("error resolving relationship to channel: %w", err)
		}

		log.Infof("resolved relationship with '%v' to channel %v", relation.Recipient.Username, channel.ID)

		if err = c.DeleteFromChannel(me, channel); err != nil {
			return err
		}
	}

	guilds, err := c.Guilds()
	if err != nil {
		return fmt.Errorf("error fetching guilds: %w", err)
	}
	for _, guild := range guilds {
		if err = c.DeleteFromGuild(me, &guild); err != nil {
			return err
		}
	}

	log.Infof("finished deleting messages: %v deleted in %v total requests", c.deletedCount, c.requestCount)

	return nil
}

func (c *Client) DeleteFromChannel(me *Me, channel *Channel) error {
	if c.skipChannel(channel.ID) {
		log.Infof("skipping message deletion for channel %v", channel.ID)
		return nil
	}

	seek := 0

	for {
		results, err := c.ChannelMessages(channel, me, &seek)
		if err != nil {
			return fmt.Errorf("error fetching messages for channel: %w", err)
		}
		if len(results.ContextMessages) == 0 {
			log.Infof("no more messages to delete for channel %v", channel.ID)
			break
		}

		if err = c.DeleteMessages(results, &seek); err != nil {
			return err
		}
	}

	return nil
}

func (c *Client) DeleteFromGuild(me *Me, channel *Channel) error {
	if c.skipChannel(channel.ID) {
		log.Infof("skipping message deletion for guild '%v'", channel.Name)
		return nil
	}

	seek := 0

	for {
		results, err := c.GuildMessages(channel, me, &seek)
		if err != nil {
			return fmt.Errorf("error fetching messages for guild: %w", err)
		}
		if len(results.ContextMessages) == 0 {
			log.Infof("no more messages to delete for guild '%v'", channel.Name)
			break
		}

		if err = c.DeleteMessages(results, &seek); err != nil {
			return err
		}
	}

	return nil
}

func (c *Client) DeleteMessages(messages *Messages, seek *int) error {
	// Milliseconds to wait between deleting messages
	// A delay which is too short will cause the server to return 429 and force us to wait a while
	// By preempting the server's delay, we can reduce the number of requests made to the server
	const minSleep = 200

	for _, ctx := range messages.ContextMessages {
		for _, msg := range ctx {
			if !msg.Hit {
				// This is a context message which may or may not be authored
				// by the current user
				log.Debugf("skipping context message")
				continue
			}

			// The message might be an action rather than text. Actions aren't deletable.
			// An example of an action is a call request.
			if msg.Type != UserMessage && msg.Type != UserReply {
				log.Debugf("found message of type %v, seeking ahead", msg.Type)
				(*seek)++
				continue
			}

			// Check if this message is in our list of channels to skip
			// This will only skip this specific message and increment the seek index
			// Entire channels should be skipped at the caller of this function
			// We do it this way because guilds searches return a mix of messages
			// from any channel
			if c.skipChannel(msg.ChannelID) {
				log.Infof("skipping message deletion for channel %v", msg.ChannelID)
				(*seek)++
				continue
			}

			if c.skipPinned && msg.Pinned {
				log.Infof("found pinned message, skipping")
				(*seek)++
				continue
			}

			log.Infof("deleting message %v from channel %v", msg.ID, msg.ChannelID)
			if c.dryRun {
				// Move seek index forward to simulate message deletion on server's side
				(*seek)++
			} else {
				if err := c.DeleteMessage(&msg); err != nil {
					return fmt.Errorf("error deleting message: %w", err)
				}
				time.Sleep(minSleep * time.Millisecond)
			}
			// Increment regardless of whether it's a dry run
			c.deletedCount++
		}
	}

	return nil
}

func (c *Client) skipChannel(channel string) bool {
	for _, skip := range c.skipChannels {
		if channel == skip {
			return true
		}
	}
	return false
}

func (c *Client) request(method string, endpoint string, reqData any, resData any) error {
	url := api + endpoint
	log.Debugf("%v %v", method, url)

	buffer := new(bytes.Buffer)
	if reqData != nil {
		if err := json.NewEncoder(buffer).Encode(reqData); err != nil {
			return fmt.Errorf("error encoding request data: %w", err)
		}
	}
	req, err := http.NewRequest(method, url, buffer)
	if err != nil {
		return fmt.Errorf("error building request: %w", err)
	}
	req.Header.Set("Authorization", c.token)
	req.Header.Set("X-Super-Properties", c.spoof.SuperProps)
	req.Header.Set("User-Agent", c.spoof.UserAgent)
	req.Header.Set("Content-Type", "application/json")

	res, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("error sending request: %w", err)
	}
	defer res.Body.Close()

	c.requestCount++

	log.Debugf("server returned status %v", http.StatusText(res.StatusCode))

	switch status := res.StatusCode; {
	case status >= http.StatusInternalServerError:
		return fmt.Errorf("bad status code %v", http.StatusText(res.StatusCode))
	case status == http.StatusAccepted:
		// retry_after is an integer in milliseconds
		if err := c.wait(res, 1); err != nil {
			return err
		}
		// Try again once we've waited for the period that the server has asked us to.
		return c.request(method, endpoint, reqData, resData)
	case status == http.StatusTooManyRequests:
		// retry_after is a float in seconds
		if err := c.wait(res, 1000); err != nil {
			return err
		}
		// Try again once we've waited for the period that the server has asked us to.
		return c.request(method, endpoint, reqData, resData)
	case status == http.StatusForbidden:
		break
	case status == http.StatusUnauthorized:
		return fmt.Errorf("bad status code %v, log out and log back in to discord or verify your token is correct", http.StatusText(res.StatusCode))
	case status == http.StatusBadRequest:
		return fmt.Errorf("bad status code %v", http.StatusText(res.StatusCode))
	case status == http.StatusNoContent:
		break
	case status == http.StatusOK:
		if err := json.NewDecoder(res.Body).Decode(resData); err != nil {
			return fmt.Errorf("error decoding response: %w", err)
		}
	default:
		return fmt.Errorf("status code %v is unhandled", http.StatusText(res.StatusCode))
	}

	return nil
}

func (c *Client) wait(res *http.Response, mult int) error {
	data := new(ServerWait)
	if err := json.NewDecoder(res.Body).Decode(data); err != nil {
		return fmt.Errorf("error decoding response: %w", err)
	}

	// Multiply retry_after by the mult passed in
	millis := time.Duration(data.RetryAfter*float32(mult)) * time.Millisecond
	log.Infof("server asked us to sleep for %v", millis)
	time.Sleep(millis)

	return nil
}

// https://discord.com/developers/docs/resources/channel#message-object-message-types
const (
	UserMessage = 0
	UserReply   = 19
)

// https://discord.com/developers/docs/resources/channel#channel-object-channel-types
const (
	DirectChannel = 1
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
	Pinned    bool   `json:"pinned"`
}

type Messages struct {
	TotalResults    int         `json:"total_results"`
	ContextMessages [][]Message `json:"messages"`
}

type ServerWait struct {
	RetryAfter float32 `json:"retry_after"`
}
