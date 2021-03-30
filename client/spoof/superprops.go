package spoof

import (
	"encoding/base64"
	"math/rand"
	"time"
)

var props = []string{
	`{"os":"Linux","browser":"Chrome","device":"","system_locale":"en-GB","browser_user_agent":"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/89.0.4389.90 Safari/537.36","browser_version":"89.0.4389.90","os_version":"","referrer":"","referring_domain":"","referrer_current":"","referring_domain_current":"","release_channel":"stable","client_build_number":80886,"client_event_source":null}`,
	`{"os":"Windows","browser":"Chrome","device":"","system_locale":"en-GB","browser_user_agent":"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/89.0.4389.90 Safari/537.36","browser_version":"89.0.4389.90","os_version":"10","referrer":"","referring_domain":"","referrer_current":"","referring_domain_current":"","release_channel":"stable","client_build_number":80886,"client_event_source":null}`,
	`{"os":"Windows","browser":"Discord Client","release_channel":"stable","client_version":"0.0.309","os_version":"10.0.19042","os_arch":"x64","system_locale":"en-GB","client_build_number":80756,"client_event_source":null}`,
}

// Generate random base64 encoded super properties object
// This makes us look more like a legit client
func RandomSuperProps() string {
	rand.Seed(time.Now().UnixNano())
	rand := rand.Intn(len(props))

	return base64.StdEncoding.EncodeToString([]byte(props[rand]))
}
