package main

import (
	"encoding/json"
	"fmt"
	"log"
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
		fmt.Println(file + ": File doesn't exist.")
		os.Exit(1)
	}

	configFile, err := os.Open(file)
	defer configFile.Close()
	if err != nil {
		fmt.Println(err.Error())
	}
	jsonParser := json.NewDecoder(configFile)
	jsonParser.Decode(&config)

	if empty(config.StatsURL) {
		log.Fatal("ERROR: Please set statsURL in configuration file.")
	}

	if config.CheckStreams {
		if empty(config.StreamPlaylistURL) {
			log.Fatal("ERROR: Please set streamPlaylistURL in configuration file.")
		}
		if empty(config.Proxy.Domain) {
			log.Fatal("ERROR: Please set proxy domain in configuration file.")
		}
		if empty(config.Proxy.SourceDomains) {
			log.Fatal("ERROR: Please set source domains in configuration file.")
		}
	}

	return config
}

func timeFormat(x time.Time, showDate bool) string {
	location, _ := time.LoadLocation("Local")
	if showDate {
		return x.In(location).Format("2006-01-02 3:04PM")
	}
	return x.In(location).Format("3:04PM")
}

func getJSON(url string, target interface{}) error {

	client := &http.Client{Timeout: 10 * time.Second}

	r, err := client.Get(url)
	if err != nil {
		return err
	}
	defer r.Body.Close()

	return json.NewDecoder(r.Body).Decode(target)
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

func checkDependencies() {

	var error error

	if !config.CheckStreams {
		return
	}

	proxyPaths := []string{"go-mlbam-proxy", "/usr/local/bin/go-mlbam-proxy"}
	for _, p := range proxyPaths {
		proxyPath, error = exec.LookPath(p)
		if error == nil {
			break
		}
	}

	if empty(proxyPath) {
		log.Fatal("ERROR: Unable to find go-mlbam-proxy in path.")
	}

	streamlinkPaths := []string{"streamlink", "/usr/local/bin/streamlink"}
	for _, p := range streamlinkPaths {
		streamlinkPath, error = exec.LookPath(p)
		if error == nil {
			break
		}
	}

	if empty(streamlinkPath) {
		log.Fatal("ERROR: Unable to find streamlink in path.")
	}

	vlcPaths := []string{"cvlc", "vlc", "/Applications/VLC.app/Contents/MacOS/VLC", "~/Applications/VLC.app/Contents/MacOS/VLC"}
	for _, p := range vlcPaths {
		vlcPath, error = exec.LookPath(p)
		if error == nil {
			break
		}
	}

	if empty(vlcPath) {
		log.Fatal("ERROR: Unable to find VLC in path.")
	}

	log.Println("streamlink =", streamlinkPath, "", "VLC =", vlcPath, "", "proxy =", proxyPath)

}
