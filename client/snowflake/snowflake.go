package snowflake

const discordEpoch = 1420070400000

func ToSnowflake(millis int64) int64 {
	return (millis - discordEpoch) << 22
}

func FromSnowflake(snowflake int64) int64 {
	return (snowflake >> 22) + discordEpoch
}
