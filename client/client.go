package client

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/cedws/discord-delete/client/snowflake"
	"github.com/cedws/discord-delete/client/spoof"

	log "github.com/sirupsen/logrus"
)

const day = time.Hour * 24

// https://discord.com/developers/docs/resources/channel#message-object-message-types
const (
	UserMessage = 0
	UserReply   = 19
)

// https://discord.com/developers/docs/resources/channel#channel-object-channel-types
const (
	DirectChannel = 1
)

var ErrorInvalidDuration = errors.New("error parsing duration")

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
	TotalResults int         `json:"total_results"`
	Messages     [][]Message `json:"messages"`
	Threads      []Thread    `json:"threads"`
}

type Thread struct {
	ID       string
	Metadata struct {
		Archived bool
		Locked   bool
	} `json:"thread_metadata"`
}

type ServerWait struct {
	RetryAfter float32 `json:"retry_after"`
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

func (c *Client) SetDryRun(dryRun bool) {
	c.dryRun = dryRun
}

func (c *Client) SetSkipChannels(skipChannels []string) {
	c.skipChannels = skipChannels
}

func (c *Client) SetSkipPinned(skipPinned bool) {
	c.skipPinned = skipPinned
}

func (c *Client) SetMinAge(minAge uint) error {
	t := time.Now().Add(-time.Duration(minAge) * day)
	millis := t.UnixNano() / int64(time.Millisecond)

	c.maxID = snowflake.ToSnowflake(millis)
	log.Debugf("message maximum ID must be %v", c.maxID)

	return nil
}

func (c *Client) SetMaxAge(maxAge uint) error {
	t := time.Now().Add(-time.Duration(maxAge) * day)
	millis := t.UnixNano() / int64(time.Millisecond)

	c.minID = snowflake.ToSnowflake(millis)
	log.Debugf("message minimum ID must be %v", c.minID)

	return nil
}

func (c *Client) Delete() error {
	me, err := c.Me()
	if err != nil {
		return fmt.Errorf("error fetching profile information: %w", err)
	}

	channels, err := c.Channels()
	if err != nil {
		return fmt.Errorf("error fetching channels: %w", err)
	}

	for _, channel := range channels {
		if err = c.DeleteFromChannel(me, channel); err != nil {
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

		channel, err := c.RelationshipChannel(relation.Recipient)
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
		if err = c.DeleteFromGuild(me, guild); err != nil {
			return err
		}
	}

	log.Infof("finished deleting messages: %v deleted in %v total requests", c.deletedCount, c.requestCount)

	return nil
}

func (c *Client) DeleteFromChannel(me Me, channel Channel) error {
	if c.skipChannel(channel.ID) {
		log.Infof("skipping message deletion for channel %v", channel.ID)
		return nil
	}

	offset := 0

	for {
		results, err := c.ChannelMessages(channel, me, offset)
		if err != nil {
			return fmt.Errorf("error fetching messages for channel: %w", err)
		}
		if len(results.Messages) == 0 {
			log.Infof("no more messages to delete for channel %v", channel.ID)
			break
		}

		if err = c.DeleteMessages(results, &offset); err != nil {
			return err
		}
	}

	return nil
}

func (c *Client) DeleteFromGuild(me Me, channel Channel) error {
	if c.skipChannel(channel.ID) {
		log.Infof("skipping message deletion for guild '%v'", channel.Name)
		return nil
	}

	offset := 0

	for {
		results, err := c.GuildMessages(channel, me, offset)
		if err != nil {
			return fmt.Errorf("error fetching messages for guild: %w", err)
		}
		if len(results.Messages) == 0 {
			log.Infof("no more messages to delete for guild '%v'", channel.Name)
			break
		}

		if err = c.DeleteMessages(results, &offset); err != nil {
			return err
		}
	}

	return nil
}

func (c *Client) DeleteMessages(messages Messages, offset *int) error {
	// Milliseconds to wait between deleting messages
	// A delay which is too short will cause the server to return 429 and force us to wait a while
	// By preempting the server's delay, we can reduce the number of requests made to the server
	const minSleep = 200

	archived := make(map[string]bool)
	for _, thread := range messages.Threads {
		archived[thread.ID] = thread.Metadata.Archived || thread.Metadata.Locked
	}

	for _, ctx := range messages.Messages {
		for _, msg := range ctx {
			if !msg.Hit {
				// message is for context but may not be authored by this user
				log.Debugf("skipping context message")
				continue
			}

			if archived[msg.ChannelID] {
				// TODO: try to unarchive the thread
				log.Debugf("message is in archived or locked thread %v", msg.ChannelID)
				(*offset)++
				continue
			}

			if msg.Type != UserMessage && msg.Type != UserReply {
				// message is not text but could be an action for example
				log.Debugf("found message of type %v, seeking ahead", msg.Type)
				(*offset)++
				continue
			}

			if c.skipPinned && msg.Pinned {
				log.Infof("found pinned message, skipping")
				(*offset)++
				continue
			}

			// Check if this message is in our list of channels to skip
			// This will only skip this specific message and increment the seek index
			// Entire channels should be skipped at the caller of this function
			// We do it this way because guilds searches return a mix of messages
			// from any channel
			if c.skipChannel(msg.ChannelID) {
				log.Infof("skipping message deletion for channel %v", msg.ChannelID)
				(*offset)++
				continue
			}

			log.Infof("deleting message %v from channel %v", msg.ID, msg.ChannelID)
			if c.dryRun {
				// Move seek index forward to simulate message deletion on server's side
				(*offset)++
			} else {
				if err := c.DeleteMessage(msg); err != nil {
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
