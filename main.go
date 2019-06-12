package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/olekukonko/tablewriter"
	"github.com/urfave/cli"
)

var (
	config  configuration
	streams map[string]Stream
	s       Schedule
)

const version = "1.0"

func hasGameStarted(state string) bool {
	switch state {
	case
		"Scheduled",
		"Postponed",
		"Pre-Game":
		return false
	}
	return true

}

func isActiveGame(state string) bool {
	switch state {
	case
		"Pre-Game",
		"Postponed",
		"Scheduled",
		"Game Over",
		"Final":
		return false
	}
	return true
}

func isCompleteGame(state string) bool {
	switch state {
	case
		"Game Over",
		"Final":
		return true
	}
	return false
}

func getTeamDisplay(teams Teams, singleLine bool) string {

	if singleLine {
		return teams.Away.Team.Name + " (" + teams.Away.Team.Abbreviation + ") vs " + teams.Home.Team.Name + " (" + teams.Home.Team.Abbreviation + ")"
	}

	return teams.Away.Team.Name + " (" + teams.Away.Team.Abbreviation + ")\n" + teams.Home.Team.Name + " (" + teams.Home.Team.Abbreviation + ")"
}

func getStreamDisplay(g Game) string {

	if len(g.Streams) == 0 || !isActiveGame(g.GameStatus.DetailedState) {
		return "\n"
	}

	var streamDisplay strings.Builder

	for _, s := range g.Streams {
		streamDisplay.WriteString(s.MediaFeedType + " (" + s.CallLetters + ") [" + s.ID + "]\n")
	}

	return strings.TrimSpace(streamDisplay.String())

}

func getGameStatusDisplay(g Game) string {

	var sd strings.Builder

	if g.GameStatus.DetailedState == "Warmup" || g.GameStatus.DetailedState == "Postponed" || g.GameStatus.DetailedState == "Final" || g.GameStatus.DetailedState == "Game Over" {
		sd.WriteString(g.GameStatus.DetailedState)
	} else if !hasGameStarted(g.GameStatus.DetailedState) {
		t, _ := time.Parse(time.RFC3339, g.GameDate)
		sd.WriteString(timeFormat(t))
	} else {
		sd.WriteString(strings.ToUpper(g.LineScore.InningState[0:3]) + " " + g.LineScore.CurrentInningOrdinal)
	}

	return sd.String()

}

func getGameScoreDisplay(g Game) string {

	scoreDisplay := "\n"

	if hasGameStarted(g.GameStatus.DetailedState) {
		scoreDisplay = strconv.Itoa(g.LineScore.Scoring.Away.Runs) + "\n" + strconv.Itoa(g.LineScore.Scoring.Home.Runs)
	}

	return scoreDisplay
}

func displayGames(s Schedule, team string) {
	table := tablewriter.NewWriter(os.Stdout)
	table.SetRowLine(true)
	table.SetColMinWidth(0, 20)

	for _, g := range s.Games {

		if team != "" && (g.Teams.Away.Team.Abbreviation != team && g.Teams.Home.Team.Abbreviation != team) {
			continue
		}

		if !isCompleteGame(g.GameStatus.DetailedState) {

			var v []string

			if config.CheckStreams && len(streams) > 0 {
				v = []string{getTeamDisplay(g.Teams, false), getGameScoreDisplay(g), getGameStatusDisplay(g), getStreamDisplay(g)}
			} else {
				v = []string{getTeamDisplay(g.Teams, false), getGameScoreDisplay(g), getGameStatusDisplay(g)}
			}

			table.Append(v)
		}

	}

	if table.NumLines() > 0 {
		table.Render()
	}
}

func refresh(team string) {
	s = GetMLBSchedule()

	if config.CheckStreams {
		checkAvailableStreams(s)
	}

	displayGames(s, team)
}

func startStream(streamID string) {

	i, ok := streams[streamID]
	if !ok {
		fmt.Println("Stream doesn't exist.")
	} else {
		runStreamlink(i)
	}
}

func exit() {
	stopProxy()
	os.Exit(0)
}

func run(c *cli.Context) {

	config = loadConfiguration(c.String("config"))
	refresh(c.String("team"))

	if !config.CheckStreams {
		exit()
	}

	startProxy()

	if len(streams) == 0 {
		fmt.Println("No streams available.")
	}

	if c.String("stream") != "" {
		startStream(c.String("stream"))
		exit()
	}

	for {

		reader := bufio.NewReader(os.Stdin)
		fmt.Print(">> ")
		input, _ := reader.ReadString('\n')

		input = strings.TrimSpace(input)

		if input == "q" {
			exit()
		} else if input == "r" {
			refresh("")
		} else if input == "h" {
			fmt.Println("[streamId] = play stream\nr = refresh\nq = quit")
		} else {
			startStream(input)
		}

	}

}

func main() {
	app := cli.NewApp()
	app.Name = "mlbme"
	app.Usage = "stream MLB games"
	app.Version = version

	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "config, c",
			Value: "config.json",
			Usage: "Load configuration from `config.json`",
		},
		cli.StringFlag{
			Name:  "team, t",
			Value: "",
			Usage: "Filter on team by abbreviation",
		},
		cli.StringFlag{
			Name:  "stream, s",
			Value: "",
			Usage: "Stream ID to stream",
		},
	}

	app.Action = func(c *cli.Context) error {
		run(c)
		return nil
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
