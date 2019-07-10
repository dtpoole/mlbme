package main

import (
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"time"
)

type configuration struct {
	StatsURL          string `json:"statsURL"`
	StreamPlaylistURL string `json:"streamPlaylistURL"`
	CheckStreams      bool   `json:"checkStreams"`
	Proxy             struct {
		Domain        string `json:"domain"`
		SourceDomains string `json:"sourceDomains"`
	} `json:"proxy"`
}

func loadConfiguration(file string) configuration {

	if _, err := os.Stat(file); os.IsNotExist(err) {
		exit(errors.New("ERROR: " + file + ": File doesn't exist"))
	}

	configFile, err := os.Open(file)
	defer configFile.Close()
	if err != nil {
		exit(err)
	}
	jsonParser := json.NewDecoder(configFile)
	jsonParser.Decode(&config)

	if empty(config.StatsURL) {
		exit(errors.New("ERROR: Please set statsURL in configuration file"))
	}

	if config.CheckStreams {
		if empty(config.StreamPlaylistURL) {
			exit(errors.New("ERROR: Please set streamPlaylistURL in configuration file"))
		}
		if empty(config.Proxy.Domain) {
			exit(errors.New("ERROR: Please set proxy domain in configuration file"))
		}
		if empty(config.Proxy.SourceDomains) {
			exit(errors.New("ERROR: Please set source domains in configuration file"))
		}
	}

	return config
}

func timeFormat(x *time.Time, showDate bool) string {
	location, _ := time.LoadLocation("Local")
	if showDate {
		return x.In(location).Format("2006-01-02 3:04PM")
	}
	return x.In(location).Format("3:04PM")
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

func httpGet(url string) *http.Response {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		exit(err)
	}
	req.Header.Set("User-Agent", UserAgent)

	resp, err := httpClient.Do(req)
	if err != nil {
		exit(err)
	}

	return resp
}

func checkDependencies() {

	var error error

	if !config.CheckStreams {
		return
	}

	proxyPaths := []string{"go-mlbam-proxy", "go-mlbam-proxy/go-mlbam-proxy", "/usr/local/bin/go-mlbam-proxy"}
	for _, p := range proxyPaths {
		if proxyPath, error = exec.LookPath(p); error == nil {
			break
		}
	}

	if empty(proxyPath) {
		exit(errors.New("ERROR: Unable to find go-mlbam-proxy in path"))
	}

	streamlinkPaths := []string{"streamlink", "/usr/local/bin/streamlink"}
	for _, p := range streamlinkPaths {
		if streamlinkPath, error = exec.LookPath(p); error == nil {
			break
		}
	}

	if empty(streamlinkPath) {
		exit(errors.New("ERROR: Unable to find streamlink in path"))
	}

	vlcPaths := []string{"cvlc", "vlc", "/Applications/VLC.app/Contents/MacOS/VLC", "~/Applications/VLC.app/Contents/MacOS/VLC"}
	for _, p := range vlcPaths {
		if vlcPath, error = exec.LookPath(p); error == nil {
			break
		}
	}

	if empty(vlcPath) {
		exit(errors.New("ERROR: Unable to find VLC in path"))
	}

}
