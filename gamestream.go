package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"sync"
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

func getGameStreams(g Game, ch chan Stream, wg *sync.WaitGroup) {

	cdns := [2]string{"akc", "l3c"}

	for _, epg := range g.Content.Media.EPG {
		if epg.Title != "MLBTV" {
			continue
		}

		for _, item := range epg.MediaItems {
			if item.MediaState == "MEDIA_ON" {

				playlist := ""

				for _, cdn := range cdns {
					streamURL := fmt.Sprintf(config.StreamPlaylistURL, schedule.Date, strconv.Itoa(item.ID), cdn)
					playlist = getM3U8Url(streamURL)
					if playlist != "" {
						break
					}
				}

				if playlist != "" {
					s := Stream{
						ID:             strconv.Itoa(item.ID),
						GamePk:         g.GamePk,
						MediaFeedType:  item.MediaFeedType,
						CallLetters:    item.CallLetters,
						StreamPlaylist: playlist,
					}

					ch <- s
				}
			}
		}
	}

	wg.Done()

}

func checkAvailableStreams() {

	var wg sync.WaitGroup

	streams = make(map[int]map[string]Stream)

	ch := make(chan Stream)

	for _, g := range *schedule.Games {
		wg.Add(1)
		go getGameStreams(g, ch, &wg)
	}

	go func() {
		wg.Wait()
		close(ch)
	}()

	for v := range ch {
		if streams[v.GamePk] == nil {
			streams[v.GamePk] = make(map[string]Stream)
		}
		streams[v.GamePk][v.ID] = v
	}

}
