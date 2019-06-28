package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/olekukonko/tablewriter"
	"github.com/urfave/cli"
)

var (
	config         configuration
	streams        map[int]map[string]Stream
	schedule       Schedule
	streamlinkPath string
	vlcPath        string
	proxyPath      string
	team           string
)

// consts
const (
	NLINE       = "\n"
	VERSION     = "1.0"
	RefreshRate = 3 * time.Minute
)

func hasGameStarted(state string) bool {
	switch state {
	case
		"Scheduled",
		"Postponed",
		"Pre-Game",
		"Warmup",
		"Delayed Start: Rain":
		return false
	}
	return true

}

func isActiveGame(state string) bool {
	switch state {
	case
		"In Progress",
		"Warmup":
		return true
	}
	return false
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

func getTeamDisplay(teams Teams) string {
	return teams.Away.Team.Name + " (" + teams.Away.Team.Abbreviation + ")\n" + teams.Home.Team.Name + " (" + teams.Home.Team.Abbreviation + ")"
}

func getStreamDisplay(g Game) string {

	if len(streams[g.GamePk]) == 0 || !isActiveGame(g.GameStatus.DetailedState) {
		return NLINE
	}

	var streamDisplay strings.Builder

	for _, s := range streams[g.GamePk] {
		streamDisplay.WriteString(s.MediaFeedType + " (" + s.CallLetters + ") [" + s.ID + "]\n")
	}

	return strings.TrimSpace(streamDisplay.String())

}

func getGameStatusDisplay(g Game) string {

	var sd strings.Builder

	if g.GameStatus.StatusCode == "I" {
		sd.WriteString(strings.ToUpper(g.LineScore.InningState[0:3]) + " " + g.LineScore.CurrentInningOrdinal)
	} else if g.GameStatus.StatusCode == "S" || g.GameStatus.StatusCode == "P" {
		t, _ := time.Parse(time.RFC3339, g.GameDate)
		sd.WriteString(timeFormat(t, false))
	} else if match(`^P*`, g.GameStatus.StatusCode) || g.GameStatus.StatusCode == "DR" {
		sd.WriteString(g.GameStatus.DetailedState)
	}

	if match("^Suspended*", g.GameStatus.DetailedState) {
		sd.WriteString("\n" + g.GameStatus.DetailedState)
	}

	return strings.TrimSpace(sd.String())

}

func getGameScoreDisplay(g Game) string {

	scoreDisplay := NLINE

	if hasGameStarted(g.GameStatus.DetailedState) {
		scoreDisplay = strconv.Itoa(g.LineScore.Scoring.Away.Runs) + NLINE + strconv.Itoa(g.LineScore.Scoring.Home.Runs)
	}

	return scoreDisplay
}

func showStreams() bool {
	if config.CheckStreams && len(streams) > 0 {
		return true
	}
	return false
}

func displayGames() {
	table := tablewriter.NewWriter(os.Stdout)
	table.SetRowLine(true)

	var v []string
	var total = 0
	showScore := false

	if schedule.CompletedGames || *schedule.TotalGamesInProgress > 0 {
		showScore = true
	}

	fmt.Println("Scoreboard: ", timeFormat(schedule.LastRefreshed, true))

	for i, g := range *schedule.Games {

		col := i % 2

		if !empty(team) && (g.Teams.Away.Team.Abbreviation != team && g.Teams.Home.Team.Abbreviation != team) {
			continue
		}

		v = append(v, getTeamDisplay(g.Teams))

		if showScore {
			v = append(v, getGameScoreDisplay(g))
		}

		v = append(v, getGameStatusDisplay(g))

		if showStreams() {
			v = append(v, getStreamDisplay(g))
		}

		if col == 0 {
			if empty(team) {
				v = append(v, NLINE)
			}
		} else {
			table.Append(v)
			v = []string{}
		}

		total = total + 1

	}

	// uneven game count
	if total == 1 && len(v) > 0 {
		table.Append(v)
	} else if len(v) > 0 {
		pad := []string{NLINE, NLINE}
		if showScore {
			pad = append(pad, NLINE)
		}
		if showStreams() {
			pad = append(pad, NLINE)
		}
		table.Append(append(v, pad...))
	}

	if table.NumLines() > 0 {
		table.Render()
	}
}

func refresh(periodic bool) {
	r := func() {
		schedule = GetMLBSchedule()
		if config.CheckStreams {
			checkAvailableStreams()
		}
	}

	if !periodic {
		r()
	} else {
		ticker := time.NewTicker(RefreshRate)
		for range ticker.C {
			r()
		}
	}
}

func startStream(streamID string, http bool) {

	var gamePk int

	for g, gs := range streams {
		if _, ok := gs[streamID]; ok {
			gamePk = g
			break
		}
	}

	if stream, ok := streams[gamePk][streamID]; ok {
		runStreamlink(stream, http)
	} else {
		fmt.Println("Stream doesn't exist.")
	}

}

func exit(code int) {
	stopProxy()
	os.Exit(code)
}

func prompt() string {

	var prompt string
	reader := bufio.NewReader(os.Stdin)

	prompt = prompt + ">> "

	fmt.Print(prompt)
	input, _ := reader.ReadString('\n')

	return strings.TrimSpace(input)
}

func run(c *cli.Context) {

	// handle ctrl-c (sigterm)
	stCh := make(chan os.Signal)
	signal.Notify(stCh, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-stCh
		exit(1)
	}()

	team = c.String("team")

	config = loadConfiguration(c.String("config"))

	checkDependencies()
	startProxy()
	refresh(false)

	// setup background refresh
	go refresh(true)

	displayGames()

	if !config.CheckStreams {
		exit(0)
	}

	if c.String("stream") != "" {
		startStream(c.String("stream"), c.Bool("http"))
		exit(0)
	}

	for {

		input := prompt()

		if input == "q" {
			exit(0)
		} else if input == "r" || input == "" {
			displayGames()
		} else if input == "h" {
			fmt.Println("[streamId] = play stream\nr = refresh\nq = quit")
		} else {
			startStream(input, c.Bool("http"))
		}

	}

}

func main() {
	app := cli.NewApp()
	app.Name = "mlbme"
	app.Usage = "stream MLB games"
	app.Version = VERSION

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
		cli.BoolFlag{
			Name:  "http",
			Usage: "Tell VLC to use HTTP streaming instead of playing locally",
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
