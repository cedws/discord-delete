package discord

type MeResponse struct {
	ID string `json:"id"`
}

type Channel struct {
	ID string `json:"id"`
}

type RelationshipResponse struct {
	ID string `json:"id"`
}

type Message struct {
	ID        string `json:"id"`
	Hit       bool   `json:"hit"`
	ChannelID string `json:"channel_id"`
	Type      int    `json:"type"`
}

type MessageContextResponse struct {
	TotalResults    int         `json:"total_results"`
	ContextMessages [][]Message `json:"messages"`
}

type TooManyRequestsResponse struct {
	RetryAfter int `json:"retry_after"`
}
