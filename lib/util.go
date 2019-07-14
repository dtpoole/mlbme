package lib

import (
	"net"
	"net/http"
	"regexp"
	"time"
)

var (
	httpClient *http.Client
)

const (
	// UserAgent to use for HTTP connections
	UserAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/75.0.3770.100 Safari/537.36"
)

func init() {
	// setup client used for all HTTP connections
	httpClient = &http.Client{
		Transport: &http.Transport{
			Dial: (&net.Dialer{
				Timeout: 5 * time.Second,
			}).Dial,
			TLSHandshakeTimeout: 5 * time.Second,
			MaxIdleConnsPerHost: 10,
		},
		Timeout: 5 * time.Second,
	}

}

func httpGet(url string) (resp *http.Response, err error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return
	}
	req.Header.Set("User-Agent", UserAgent)

	resp, err = httpClient.Do(req)
	if err != nil {
		return
	}
	return
}

func empty(in string) bool {
	if in == "" {
		return true
	}
	return false
}

func match(pattern string, in string) bool {
	m, _ := regexp.MatchString(pattern, in)
	return m
}

func timeFormat(x *time.Time, showDate bool) string {
	location, _ := time.LoadLocation("Local")
	if showDate {
		return x.In(location).Format("2006-01-02 3:04PM")
	}
	return x.In(location).Format("3:04PM")
}

func hasGameStarted(state string) bool {
	switch state {
	case
		"Scheduled",
		"Postponed",
		"Pre-Game",
		"Warmup",
		"Delayed Start: Rain",
		"Delayed Start: Lightning":
		return false
	}
	return true

}

func isActiveGame(state string) bool {
	switch state {
	case
		"In Progress",
		"Warmup",
		"Delayed: Rain":
		return true
	}
	return false
}

func isDelayedSuspended(state string) bool {
	switch state {
	case
		"Ceremony",
		"Suspended: Rain",
		"Delayed: Rain":
		return true
	}
	return false

}

func isCompleteGame(state string) bool {
	switch state {
	case
		"Game Over",
		"Final":
		return true
	}
	return false
}
