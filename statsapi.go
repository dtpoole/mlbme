package main

import (
	"fmt"
	"time"
)

type Score struct {
	Runs   int `json:"runs"`
	Hits   int `json:"hits"`
	Errors int `json:"errors"`
}

type Scoring struct {
	Home Score `json:"home"`
	Away Score `json:"away"`
}

type LineScore struct {
	CurrentInning        int     `json:"currentInning"`
	CurrentInningOrdinal string  `json:"currentInningOrdinal"`
	InningState          string  `json:"inningState"`
	Scoring              Scoring `json:"teams"`
}

type Team struct {
	Name         string `json:"teamName"`
	FullName     string `json:"name"`
	Abbreviation string `json:"abbreviation"`
}

type Teams struct {
	Away struct {
		Team Team `json:"team"`
	} `json:"away"`
	Home struct {
		Team Team `json:"team"`
	} `json:"home"`
}

type MediaItem struct {
	ID            int    `json:"id"`
	MediaID       string `json:"mediaId"`
	MediaState    string `json:"mediaState"`
	ContentID     string `json:"contentId"`
	MediaFeedType string `json:"mediaFeedType"`
	CallLetters   string `json:"callLetters"`
}

type Game struct {
	GamePk     int   `json:"gamePk"`
	Teams      Teams `json:"teams"`
	GameStatus struct {
		DetailedState string `json:"detailedState"`
	} `json:"status"`
	GameDate string `json:"gameDate"`
	Content  struct {
		Media struct {
			EPG []struct {
				Title      string      `json:"title"`
				MediaItems []MediaItem `json:"items"`
			} `json:"epg"`
		} `json:"media"`
		Link string `json:"link"`
	} `json:"content"`
	LineScore LineScore `json:"linescore"`
}

type Data struct {
	Dates []struct {
		TotalGames           int    `json:"totalGames"`
		TotalGamesInProgress int    `json:"totalGamesInProgress"`
		Games                []Game `json:"games"`
	} `json:"dates"`
}

type Schedule struct {
	Date                 string
	URL                  string
	TotalGames           int
	TotalGamesInProgress int
	Games                []Game
}

func GetMLBSchedule() Schedule {

	var s = new(Schedule)

	// check for in progress games after midnight, but before 3AM
	dt := time.Now()
	if dt.Hour() <= 3 {
		s.Date = time.Now().AddDate(0, 0, -1).Local().Format("2006-01-02")
	} else {
		s.Date = dt.Local().Format("2006-01-02")
	}

	s.URL = fmt.Sprintf(config.StatsURL, s.Date)

	d := new(Data)
	getJSON(s.URL, &d)

	s.TotalGames = d.Dates[0].TotalGames
	s.TotalGamesInProgress = d.Dates[0].TotalGamesInProgress
	s.Games = d.Dates[0].Games

	return *s
}
