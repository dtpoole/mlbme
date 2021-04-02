package lib

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
)

// Config hold app configuration options
type Config struct {
	StatsURL          string `json:"statsURL"`
	StreamPlaylistURL string `json:"streamPlaylistURL"`
	CheckStreams      bool   `json:"checkStreams"`
	Proxy             struct {
		Domain        string `json:"domain"`
		SourceDomains string `json:"sourceDomains"`
	} `json:"proxy"`
}

// LoadConfig - load configuration from JSON file
func LoadConfig(file string) (config *Config, err error) {

	if _, err = os.Stat(file); os.IsNotExist(err) {
		err = fmt.Errorf("file %s not found", file)
		return
	}

	configFile, err := os.Open(file)

	if err != nil {
		err = fmt.Errorf("unable to open %s", file)
		return
	}

	jsonParser := json.NewDecoder(configFile)
	if err = jsonParser.Decode(&config); err != nil {
		err = errors.New("unable to parse JSON config file")
		return
	}

	configFile.Close()

	if config.StatsURL == "" {
		err = errors.New("set statsURL in configuration file")
	}

	if config.CheckStreams {
		if config.StreamPlaylistURL == "" {
			err = errors.New("set streamPlaylistURL in configuration file")
		}
		if config.Proxy.Domain == "" {
			err = errors.New("set proxy domain in configuration file")
		}
		if config.Proxy.SourceDomains == "" {
			err = errors.New("set source domains in configuration file")
		}
	}

	return
}
