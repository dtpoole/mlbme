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
		err = fmt.Errorf("File %s not found", file)
		return
	}

	configFile, err := os.Open(file)
	defer configFile.Close()
	if err != nil {
		err = fmt.Errorf("Unable to open %s", file)
		return
	}

	jsonParser := json.NewDecoder(configFile)
	if err = jsonParser.Decode(&config); err != nil {
		err = errors.New("Unable to parse JSON config file")
		return
	}

	if config.StatsURL == "" {
		err = errors.New("Set statsURL in configuration file")
	}

	if config.CheckStreams {
		if config.StreamPlaylistURL == "" {
			err = errors.New("Set streamPlaylistURL in configuration file")
		}
		if config.Proxy.Domain == "" {
			err = errors.New("Set proxy domain in configuration file")
		}
		if config.Proxy.SourceDomains == "" {
			err = errors.New("Set source domains in configuration file")
		}
	}

	return
}
