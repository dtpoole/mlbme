package lib

import (
	"fmt"
	"io/ioutil"
	"regexp"
	"strconv"
	"sync"

	log "github.com/sirupsen/logrus"
)

// GameStreams struct holds game stream information
type GameStreams struct {
	config   *Config
	schedule *Schedule
	Streams  map[int]map[string]*Stream
}

// Stream contains information on the video stream.
type Stream struct {
	GamePk                     int
	ID, StreamPlaylist         string
	MediaFeedType, CallLetters string
}

// NewGameStreams creates a GameStreams
func NewGameStreams(c *Config, s *Schedule) (gs GameStreams) {
	gs.config = c
	gs.schedule = s
	gs.Streams = make(map[int]map[string]*Stream)

	return
}

func (gs *GameStreams) getPlaylistURL(url string) (playlist string, err error) {

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

	log.WithFields(log.Fields{
		"url":          url,
		"responseData": string(responseData),
	}).Debug("stream playlist")

	// stream not available
	notAvailable, err = regexp.Match(`^Not*`, responseData)
	if notAvailable || err != nil {
		return
	}

	playlist = string(responseData)

	return

}

func (gs *GameStreams) findGameStreams(g Game, ch chan *Stream, wg *sync.WaitGroup) {

	cdns := [2]string{"akc", "l3c"}

	for _, epg := range g.Content.Media.EPG {
		if epg.Title != "MLBTV" {
			continue
		}

		for _, item := range epg.MediaItems {

			log.WithFields(log.Fields{
				"streamID":      item.ID,
				"gamePK":        g.GamePk,
				"mediaState":    item.MediaState,
				"mediaFeedType": item.MediaFeedType,
				"callLetters":   item.CallLetters,
			}).Debug("Found media item")

			if item.MediaState == "MEDIA_ON" {

				playlist := ""

				for _, cdn := range cdns {
					streamURL := fmt.Sprintf(gs.config.StreamPlaylistURL, gs.schedule.Date, strconv.Itoa(item.ID), cdn)
					playlist, _ = gs.getPlaylistURL(streamURL)
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

					log.WithFields(log.Fields{
						"streamID":       item.ID,
						"gamePK":         g.GamePk,
						"mediaState":     item.MediaState,
						"mediaFeedType":  item.MediaFeedType,
						"callLetters":    item.CallLetters,
						"streamPlaylist": playlist,
					}).Debug("Found stream")

				}
			}
		}
	}

	wg.Done()

}

// GetAvailableStreams check for available game streams
func (gs *GameStreams) GetAvailableStreams() {

	var wg sync.WaitGroup
	ch := make(chan *Stream)

	log.Debug("Checking for game streams")

	for _, g := range *gs.schedule.Games {
		wg.Add(1)
		go gs.findGameStreams(g, ch, &wg)
	}

	go func() {
		wg.Wait()
		close(ch)
	}()

	for v := range ch {
		if gs.Streams[v.GamePk] == nil {
			gs.Streams[v.GamePk] = make(map[string]*Stream)
		}
		gs.Streams[v.GamePk][v.ID] = v
	}

	log.WithFields(log.Fields{
		"streamCnt": len(gs.Streams),
	}).Debug("Finished checking streams")

	return

}
