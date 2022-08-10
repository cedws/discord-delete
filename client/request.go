package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
)

const (
	api          = "https://discord.com/api/v10"
	messageLimit = 25
)

type RequestArgs struct {
	IncludeNSFW bool
	AuthorID    string
	MinID       int64
	MaxID       int64
	Offset      int
	Limit       int
}

func (r RequestArgs) MarshalText() string {
	var args []string
	if r.IncludeNSFW {
		args = append(args, "include_nsfw=true")
	}
	if r.AuthorID != "" {
		args = append(args, fmt.Sprintf("author_id=%v", r.AuthorID))
	}
	if r.MinID != 0 {
		args = append(args, fmt.Sprintf("min_id=%v", r.MinID))
	}
	if r.MaxID != 0 {
		args = append(args, fmt.Sprintf("max_id=%v", r.MaxID))
	}
	if r.Offset != 0 {
		args = append(args, fmt.Sprintf("offset=%v", r.Offset))
	}
	if r.Limit != 0 {
		args = append(args, fmt.Sprintf("limit=%v", r.Limit))
	}

	return "?" + strings.Join(args, "&")
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

	_, err = io.Copy(io.Discard, res.Body)
	return err
}

func (c *Client) wait(res *http.Response, mult int) error {
	data := new(ServerWait)
	if err := json.NewDecoder(res.Body).Decode(data); err != nil {
		return fmt.Errorf("error decoding response: %w", err)
	}

	millis := time.Duration(data.RetryAfter*float32(mult)) * time.Millisecond
	log.Infof("server asked us to sleep for %v", millis)
	time.Sleep(millis)

	return nil
}

func (c *Client) DeleteMessage(msg Message) (err error) {
	endpoint := fmt.Sprintf(
		"/channels/%v/messages/%v",
		msg.ChannelID,
		msg.ID,
	)
	err = c.request("DELETE", endpoint, nil, nil)
	return
}

func (c *Client) Me() (me Me, err error) {
	err = c.request("GET", "/users/@me", nil, &me)
	return
}

func (c *Client) Channels() (channels []Channel, err error) {
	err = c.request("GET", "/users/@me/channels", nil, &channels)
	return
}

func (c *Client) ChannelMessages(channel Channel, me Me, offset int) (messages Messages, err error) {
	endpoint := fmt.Sprintf(
		"/channels/%v/messages/search",
		channel.ID,
	)
	args := RequestArgs{
		IncludeNSFW: true,
		AuthorID:    me.ID,
		Offset:      offset,
		Limit:       messageLimit,
		MinID:       c.minID,
		MaxID:       c.maxID,
	}

	err = c.request("GET", endpoint+args.MarshalText(), nil, &messages)
	return
}

func (c *Client) RelationshipChannel(relation Recipient) (channel Channel, err error) {
	recipients := struct {
		Recipients []string `json:"recipients"`
	}{
		[]string{relation.ID},
	}

	err = c.request("POST", "/users/@me/channels", recipients, &channel)
	return
}

func (c *Client) Relationships() (relations []Relationship, err error) {
	err = c.request("GET", "/users/@me/relationships", nil, &relations)
	return
}

func (c *Client) Guilds() (channels []Channel, err error) {
	err = c.request("GET", "/users/@me/guilds", nil, &channels)
	return
}

func (c *Client) GuildMessages(channel Channel, me Me, offset int) (messages Messages, err error) {
	endpoint := fmt.Sprintf(
		"/guilds/%v/messages/search",
		channel.ID,
	)
	args := RequestArgs{
		IncludeNSFW: true,
		AuthorID:    me.ID,
		Offset:      offset,
		Limit:       messageLimit,
		MinID:       c.minID,
		MaxID:       c.maxID,
	}

	err = c.request("GET", endpoint+args.MarshalText(), nil, &messages)
	return
}
