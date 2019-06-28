package main

import (
	"fmt"
	"time"
)

// Score represents line score for team
type Score struct {
	Runs int `json:"runs"`
}

// Scoring holds Home and Away team line score
type Scoring struct {
	Home Score `json:"home"`
	Away Score `json:"away"`
}

// LineScore contains information about the current state of the game
type LineScore struct {
	CurrentInningOrdinal string  `json:"currentInningOrdinal"`
	InningState          string  `json:"inningState"`
	Scoring              Scoring `json:"teams"`
}

// Team details
type Team struct {
	Name         string `json:"teamName"`
	Abbreviation string `json:"abbreviation"`
}

// Teams contains the teams playing
type Teams struct {
	Away struct {
		Team Team `json:"team"`
	} `json:"away"`
	Home struct {
		Team Team `json:"team"`
	} `json:"home"`
}

// MediaItem contains media available for the game
type MediaItem struct {
	ID            int    `json:"id"`
	MediaState    string `json:"mediaState"`
	MediaFeedType string `json:"mediaFeedType"`
	CallLetters   string `json:"callLetters"`
}

// Game has details on a game.
type Game struct {
	GamePk     int   `json:"gamePk"`
	Teams      Teams `json:"teams"`
	GameStatus struct {
		DetailedState string `json:"detailedState"`
		StatusCode    string `json:"statusCode"`
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

// Data root of the JSON. Contains Dates array.
type Data struct {
	Dates []struct {
		TotalGames           int    `json:"totalGames"`
		TotalGamesInProgress int    `json:"totalGamesInProgress"`
		Games                []Game `json:"games"`
	} `json:"dates"`
}

// Schedule contains details of the day's MLB games.
type Schedule struct {
	Date                 string
	URL                  string
	TotalGames           *int
	TotalGamesInProgress *int
	CompletedGames       bool
	Games                *[]Game
	LastRefreshed        time.Time
}

// GetMLBSchedule gets today's schedule of games
func GetMLBSchedule() Schedule {

	var s = new(Schedule)

	// check for in progress games after midnight, but before 3AM
	dt := time.Now()
	if dt.Hour() <= 3 {
		s.Date = time.Now().AddDate(0, 0, -1).Local().Format("2006-01-02")
	} else {
		s.Date = dt.Local().Format("2006-01-02")
	}

	s.LastRefreshed = dt

	// we are only getting one day of data
	s.URL = fmt.Sprintf(config.StatsURL, s.Date)

	d := new(Data)
	getJSON(s.URL, &d)

	s.TotalGames = &d.Dates[0].TotalGames
	s.TotalGamesInProgress = &d.Dates[0].TotalGamesInProgress
	s.Games = &d.Dates[0].Games

	for _, g := range *s.Games {
		if isCompleteGame(g.GameStatus.DetailedState) {
			s.CompletedGames = true
			break
		}
	}

	return *s
}
