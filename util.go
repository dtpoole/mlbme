package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"
)

type configuration struct {
	StatsURL          string `json:"statsURL"`
	StreamPlaylistURL string `json:"streamPlaylistURL"`
	CDN               string `json:"cdn"`
	CheckStreams      bool   `json:"checkStreams"`
	Proxy             struct {
		Path          string `json:"path"`
		Port          int    `json:"port"`
		Domain        string `json:"domain"`
		SourceDomains string `json:"sourceDomains"`
	} `json:"proxy"`
	VLC struct {
		Path string `json:"path"`
		Args string `json:"args"`
	} `json:"vlc"`
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
