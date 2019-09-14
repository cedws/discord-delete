package discord

import (
	"fmt"
	"io/ioutil"
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
	"channel_msgs": "/channels/{}/messages/search" +
		"?author_id={}" +
		"&include_nsfw=true" +
		"&offset={}" +
		"&limit={}",
	"delete_msg": "/channels/{}/messages/{}",
}

type Client struct {
	Token string
}

func (d Client) Me() string {
	endpoint := fmt.Sprintf("%v%v", api, endpoints["me"])

	client := &http.Client{}
	req, _ := http.NewRequest("GET", endpoint, nil)
	req.Header.Set("Authorization", d.Token)
	res, _ := client.Do(req)

	body, _ := ioutil.ReadAll(res.Body)
	fmt.Printf(string(body))
	return ""
}
