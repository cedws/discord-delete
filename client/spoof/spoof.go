package spoof

import (
	"encoding/base64"
	"fmt"
	"math/rand"
	"time"
)

type Info struct {
	SuperProps string
	UserAgent  string
}

var useragents = []string{
	`Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/89.0.4389.90 Safari/537.36`,
	`Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/89.0.4389.90 Safari/537.36`,
}

var superprops = []string{
	`{"os":"Linux","browser":"Chrome","device":"","system_locale":"en-GB","browser_user_agent":"%v","browser_version":"89.0.4389.90","os_version":"","referrer":"","referring_domain":"","referrer_current":"","referring_domain_current":"","release_channel":"stable","client_build_number":80886,"client_event_source":null}`,
	`{"os":"Windows","browser":"Chrome","device":"","system_locale":"en-GB","browser_user_agent":"%v","browser_version":"89.0.4389.90","os_version":"10","referrer":"","referring_domain":"","referrer_current":"","referring_domain_current":"","release_channel":"stable","client_build_number":80886,"client_event_source":null}`,
}

func RandomInfo() Info {
	// Pick random user agent
	rand.Seed(time.Now().UnixNano())
	idx := rand.Intn(len(useragents))
	ua := useragents[idx]

	// Pick random super props object
	rand.Seed(time.Now().UnixNano())
	idx = rand.Intn(len(useragents))
	sp := superprops[idx]

	// Generate random base64 encoded super properties object
	// This makes us look more like a legit client
	sprops := base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf(sp, ua)))

	return Info{
		sprops,
		ua,
	}
}
