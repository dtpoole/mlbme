package main

import (
	"bufio"
	"flag"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/olekukonko/tablewriter"
)

var (
	config                             configuration
	streams                            map[int]map[string]Stream
	schedule                           Schedule
	streamlinkPath, vlcPath, proxyPath string
	httpClient                         *http.Client
)

var (
	configFlag = flag.String("config", "config.json", "JSON configuration")
	httpFlag   = flag.Bool("http", false, "Tell VLC to use HTTP streaming instead of playing locally")
	teamFlag   = flag.String("team", "", "Filter on team by abbreviation")
	streamFlag = flag.String("stream", "", "Call letter of stream to start")
)

// consts
const (
	NLINE       = "\n"
	VERSION     = "1.0"
	RefreshRate = 3 * time.Minute
	UserAgent   = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/75.0.3770.100 Safari/537.36"
)

func init() {

	httpClient = &http.Client{
		Transport: &http.Transport{
			Dial: (&net.Dialer{
				Timeout: 5 * time.Second,
			}).Dial,
			TLSHandshakeTimeout: 5 * time.Second,
			MaxIdleConnsPerHost: 10,
		},
		Timeout: 5 * time.Second,
	}

	// handle ctrl-c (sigterm)
	stCh := make(chan os.Signal)
	signal.Notify(stCh, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-stCh
		exit(1)
	}()
}

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

func getTeamDisplay(g *Game, singleLine bool) string {

	delim := NLINE
	if singleLine {
		delim = " vs "
	}
	return g.Teams.Away.Team.Name + " (" + g.Teams.Away.Team.Abbreviation + ")" + delim + g.Teams.Home.Team.Name + " (" + g.Teams.Home.Team.Abbreviation + ")"
}

func getStreamDisplay(g *Game) string {

	if len(streams[g.GamePk]) == 0 {
		return NLINE
	}

	var streamDisplay strings.Builder

	for _, s := range streams[g.GamePk] {
		streamDisplay.WriteString(s.MediaFeedType + " [" + s.CallLetters + "]\n")
	}

	return strings.TrimSpace(streamDisplay.String())

}

func getGameStatusDisplay(g *Game) string {

	sd := &strings.Builder{}

	if g.GameStatus.StatusCode == "I" {
		sd.WriteString(strings.ToUpper(g.LineScore.InningState[0:3]) + " " + g.LineScore.CurrentInningOrdinal)
	} else if g.GameStatus.StatusCode == "S" || g.GameStatus.StatusCode == "P" {
		t, _ := time.Parse(time.RFC3339, g.GameDate)
		sd.WriteString(timeFormat(&t, false))
	} else if match(`^P*`, g.GameStatus.StatusCode) || g.GameStatus.StatusCode == "DR" {
		sd.WriteString(g.GameStatus.DetailedState)
	}

	if match("^Suspended*", g.GameStatus.DetailedState) {
		sd.WriteString("\n" + g.GameStatus.DetailedState)
	}

	return strings.TrimSpace(sd.String())

}

func getGameScoreDisplay(g *Game) string {

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

func generateGameTable() string {

	ts := &strings.Builder{}
	table := tablewriter.NewWriter(ts)
	table.SetRowLine(true)

	var v []string
	var total = 0
	showScore := false

	if schedule.CompletedGames || *schedule.TotalGamesInProgress > 0 {
		showScore = true
	}

	fmt.Println("Scoreboard for", schedule.Date+"\t\t", "("+timeFormat(&schedule.LastRefreshed, true)+")")

	for i, g := range *schedule.Games {

		col := i % 2

		if !empty(*teamFlag) && (g.Teams.Away.Team.Abbreviation != *teamFlag && g.Teams.Home.Team.Abbreviation != *teamFlag) {
			continue
		}

		v = append(v, getTeamDisplay(&g, false))

		if showScore {
			v = append(v, getGameScoreDisplay(&g))
		}

		v = append(v, getGameStatusDisplay(&g))

		if showStreams() {
			v = append(v, getStreamDisplay(&g))
		}

		if col == 0 {
			if empty(*teamFlag) {
				v = append(v, NLINE)
			}
		} else {
			table.Append(v)
			v = []string{}
		}

		total = total + 1

	}

	if total == 0 {
		ts.WriteString("No Games\n")
		// uneven game count
	} else if total == 1 && len(v) > 0 {
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

	if total > 0 && config.CheckStreams && len(streams) == 0 {
		ts.WriteString("No streams available.\n")
	}

	return ts.String()
}

func generateStreamTable(streams []Stream) string {

	ts := &strings.Builder{}
	table := tablewriter.NewWriter(ts)
	table.SetAutoMergeCells(true)
	table.SetRowLine(true)
	table.SetColMinWidth(1, 50)

	ts.WriteString("Multiple Games for " + streams[0].MediaFeedType + " [" + streams[0].CallLetters + "]...")
	for _, s := range streams {
		g := schedule.GameMap[s.GamePk]
		table.Append([]string{s.ID, getTeamDisplay(&g, true)})
	}
	table.Render()
	return ts.String()

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

	var strs []Stream

	for _, gs := range streams {
		for _, s := range gs {
			if s.ID == streamID || s.CallLetters == streamID {
				strs = append(strs, s)
			}
		}
	}

	switch len(strs) {
	case 0:
		fmt.Println("Stream doesn't exist.")
	case 1:
		runStreamlink(strs[0], http)
	default:
		fmt.Println(generateStreamTable(strs))
	}
}

func exit(code int) {
	stopStreamlink()
	stopProxy()
	os.Exit(code)
}

func prompt() string {

	reader := bufio.NewReader(os.Stdin)

	fmt.Print(">> ")
	input, _ := reader.ReadString('\n')

	return strings.ToUpper(strings.TrimSpace(input))
}

func main() {

	flag.Parse()

	config = loadConfiguration(*configFlag)

	checkDependencies()
	startProxy()
	refresh(false)

	// setup background refresh
	go refresh(true)

	fmt.Print(generateGameTable())

	if *streamFlag != "" {
		startStream(*streamFlag, *httpFlag)
	}

	for {
		input := prompt()
		if input == "Q" {
			exit(0)
		} else if input == "R" || input == "" {
			fmt.Println(generateGameTable())
		} else if input == "H" {
			fmt.Println("[call letters] = play stream\nr = refresh\nq = quit")
		} else {
			startStream(input, *httpFlag)
		}
	}
}
