package client

const discordEpoch = 1420070400000

func toSnowflake(millis int64) int64 {
	return (millis - discordEpoch) << 22
}

func fromSnowflake(snowflake int64) int64 {
	return (snowflake >> 22) + discordEpoch
}
