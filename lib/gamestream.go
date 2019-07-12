package lib

import (
	"fmt"
	"io/ioutil"
	"regexp"
	"strconv"
	"strings"
	"sync"
)

// Stream contains information on the video stream.
type Stream struct {
	GamePk                     int
	ID, StreamPlaylist         string
	MediaFeedType, CallLetters string
}

func getPlaylistURL(url string) (playlist string, err error) {

	notAvailable := false

	resp, err := httpGet(url)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	responseData, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}

	// stream not available
	notAvailable, err = regexp.Match(`^Not*`, responseData)
	if notAvailable || err != nil {
		return
	}

	// rewrite https to http
	// TODO: check if needed
	playlist = strings.Replace(string(responseData), "https", "http", 1)

	return

}

func getGameStreams(g Game, ch chan *Stream, wg *sync.WaitGroup, streamPlaylistURL string) {

	cdns := [2]string{"akc", "l3c"}

	for _, epg := range g.Content.Media.EPG {
		if epg.Title != "MLBTV" {
			continue
		}

		for _, item := range epg.MediaItems {
			if item.MediaState == "MEDIA_ON" {

				playlist := ""

				for _, cdn := range cdns {
					streamURL := fmt.Sprintf(streamPlaylistURL, g.GameDate, strconv.Itoa(item.ID), cdn)
					playlist, _ = getPlaylistURL(streamURL)
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

					ch <- &s
				}
			}
		}
	}

	wg.Done()

}

// GetAvailableStreams check for available game streams
func GetAvailableStreams(c *Config, s *Schedule) (streams map[int]map[string]*Stream) {

	var wg sync.WaitGroup

	//streams = make(map[int]map[string]Stream)

	ch := make(chan *Stream)

	for _, g := range *s.Games {
		wg.Add(1)
		go getGameStreams(g, ch, &wg, c.StreamPlaylistURL)
	}

	go func() {
		wg.Wait()
		close(ch)
	}()

	for v := range ch {
		if streams[v.GamePk] == nil {
			streams[v.GamePk] = make(map[string]*Stream)
		}
		streams[v.GamePk][v.ID] = v
	}

	return

}
