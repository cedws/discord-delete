package client

const discordEpoch = 1420070400000

func ToSnowflake(millis uint64) uint64 {
	return (millis - discordEpoch) << 22
}

func FromSnowflake(snowflake uint64) uint64 {
	return (snowflake >> 22) + discordEpoch
}
