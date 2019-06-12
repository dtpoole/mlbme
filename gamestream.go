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

func getStreamPlaylist(date string, id int) string {

	cdns := [2]string{"akc", "l3c"}

	for _, cdn := range cdns {
		streamURL := fmt.Sprintf(config.StreamPlaylistURL, date, strconv.Itoa(id), cdn)
		playlist := getM3U8Url(streamURL)
		if playlist != "" {
			return playlist
		}
	}
	return ""
}

func checkAvailableStreams(s Schedule) {

	streams = make(map[string]Stream)

	for i, g := range s.Games {
		gs := make(map[string]Stream)

		if !isActiveGame(g.GameStatus.DetailedState) {
			continue
		}

		for _, epg := range g.Content.Media.EPG {
			if epg.Title != "MLBTV" {
				continue
			}

			for _, item := range epg.MediaItems {
				if item.MediaState == "MEDIA_ON" {

					streamPlaylist := getStreamPlaylist(s.Date, item.ID)

					if streamPlaylist != "" {
						var si = new(Stream)
						si.ID = strconv.Itoa(item.ID)
						si.MediaFeedType = item.MediaFeedType
						si.CallLetters = item.CallLetters
						si.StreamPlaylist = streamPlaylist

						gs[si.ID] = *si
						streams[si.ID] = *si
					}
				}

			}
		}

		s.Games[i].Streams = gs

	}

}
