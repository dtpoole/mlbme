package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"
)

// StreamInfo contains information on the video stream.
type StreamInfo struct {
	StreamAvailable bool
	StreamURL       string
	StreamPlaylist  string
	MediaItem       *MediaItem
}

func getM3U8Url(url string) string {

	response, err := http.Get(url)
	if err != nil {
		log.Fatal(err)
	}
	defer response.Body.Close()

	responseData, err := ioutil.ReadAll(response.Body)
	if err != nil {
		log.Fatal(err)
	}

	// stream not available
	matched, err := regexp.Match(`^Not*`, responseData)
	if err != nil {
		log.Fatal(err)
	}

	if matched {
		return ""
	}

	return strings.Replace(string(responseData), "https", "http", 1)

}

func checkAvailableStreams(s Schedule) map[string]StreamInfo {

	m := make(map[string]StreamInfo)

	for _, g := range s.Games {

		if !isActiveGame(g.GameStatus.DetailedState) {
			continue
		}

		for _, epg := range g.Content.Media.EPG {
			if epg.Title != "MLBTV" {
				continue
			}

			for _, item := range epg.MediaItems {

				var si = new(StreamInfo)
				si.StreamAvailable = false
				si.MediaItem = &item

				if item.MediaState == "MEDIA_ON" {
					si.StreamURL = fmt.Sprintf(config.StreamPlaylistURL, s.Date, strconv.Itoa(item.ID), config.CDN)
					if si.StreamURL != "" {
						si.StreamPlaylist = getM3U8Url(si.StreamURL)
						if si.StreamPlaylist != "" {
							si.StreamAvailable = true
						}
					}
				}

				if si.StreamAvailable {
					m[strconv.Itoa(item.ID)] = *si
				}
			}
		}
	}

	return m

}
