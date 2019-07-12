package lib

import (
	"encoding/json"
	"os"

	"github.com/pkg/errors"
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
func LoadConfig(file string) (config Config, err error) {

	if _, err = os.Stat(file); os.IsNotExist(err) {
		err = errors.Wrapf(err,
			"File %s not found",
			file)
		return
	}

	configFile, err := os.Open(file)
	defer configFile.Close()
	if err != nil {
		err = errors.Wrapf(err,
			"Unable to open %s",
			file)
		return
	}

	jsonParser := json.NewDecoder(configFile)
	jsonParser.Decode(&config)

	if config.StatsURL == "" {
		err = errors.New("Please set statsURL in configuration file")
	}

	if config.CheckStreams {
		if config.StreamPlaylistURL == "" {
			err = errors.New("Please set streamPlaylistURL in configuration file")
		}
		if config.Proxy.Domain == "" {
			err = errors.New("Please set proxy domain in configuration file")
		}
		if config.Proxy.SourceDomains == "" {
			err = errors.New("Please set source domains in configuration file")
		}
	}

	return
}
