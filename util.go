package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"time"
)

type configuration struct {
	StatsURL          string `json:"statsURL"`
	StreamPlaylistURL string `json:"streamPlaylistURL"`
	CheckStreams      bool   `json:"checkStreams"`
	Proxy             struct {
		Path          string `json:"path"`
		Port          int    `json:"port"`
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
	return config
}

func timeFormat(x time.Time) string {
	location, _ := time.LoadLocation("Local")
	return x.In(location).Format("2006-01-02 3:04PM")
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

func checkDependencies() {

	var error error

	proxyPath, error = exec.LookPath("go-mlbam-proxy")
	if error != nil {
		log.Fatal("ERROR: Unable to find proxy.")
	}

	streamlinkPath, error = exec.LookPath("streamlink")
	if error != nil {
		log.Fatal("ERROR: Unable to find streamlink.")
	}

	vlcPaths := []string{"cvlc", "vlc", "/Applications/VLC.app/Contents/MacOS/VLC", "~/Applications/VLC.app/Contents/MacOS/VLC"}

	for _, p := range vlcPaths {
		vlcPath, error = exec.LookPath(p)
		if error == nil {
			break
		}
	}

	if vlcPath == "" {
		log.Fatal("ERROR: Unable to find VLC.")
	}

	log.Println("Using: streamlink =", streamlinkPath, "", "VLC =", vlcPath, "", "proxy =", proxyPath)

}
