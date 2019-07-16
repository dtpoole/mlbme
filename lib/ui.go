package lib

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/olekukonko/tablewriter"
)

const (
	nl = "\n"
)

// UI struct
type UI struct {
	config   *Config
	schedule *Schedule
	streams  map[int]map[string]*Stream
	team     string
}

// NewUI creates the UI struct
func NewUI(c *Config, s *Schedule, streams map[int]map[string]*Stream, team string) (ui UI) {
	ui.config = c
	ui.schedule = s
	ui.streams = streams
	ui.team = team
	return
}

// GenerateScoreboard builds the scoreboard display
func (ui *UI) GenerateScoreboard() string {

	ts := &strings.Builder{}
	table := tablewriter.NewWriter(ts)
	table.SetRowLine(true)

	var v []string
	var total = 0
	showScore := false
	showStreams := ui.showStreams()

	if ui.schedule.CompletedGames || ui.schedule.InProgressGames {
		showScore = true
	}

	fmt.Println("------\nScoreboard for", ui.schedule.Date, "(as of "+timeFormat(&ui.schedule.LastRefreshed, false)+")")

	for i, g := range *ui.schedule.Games {

		col := i % 2

		if !empty(ui.team) && (g.Teams.Away.Team.Abbreviation != ui.team && g.Teams.Home.Team.Abbreviation != ui.team) {
			continue
		}

		v = append(v, ui.getTeamDisplay(&g, false))

		if showScore {
			v = append(v, ui.getGameScoreDisplay(&g))
		}

		v = append(v, ui.getGameStatusDisplay(&g))

		if showStreams {
			v = append(v, ui.getStreamDisplay(&g))
		}

		if col == 0 {
			if empty(ui.team) {
				v = append(v, nl)
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
		pad := []string{nl, nl}
		if showScore {
			pad = append(pad, nl)
		}
		if showStreams {
			pad = append(pad, nl)
		}
		table.Append(append(v, pad...))
	}

	if table.NumLines() > 0 {
		table.Render()
	}

	if total > 0 && ui.config.CheckStreams && len(ui.streams) == 0 {
		ts.WriteString("No streams available.\n------\n")
	}

	return ts.String()

}

func (ui *UI) showStreams() bool {
	if ui.config.CheckStreams && len(ui.streams) > 0 {
		return true
	}
	return false
}

func (ui *UI) getStreamDisplay(g *Game) (s string) {

	if len(ui.streams[g.GamePk]) == 0 {
		return nl
	}

	var streamDisplay strings.Builder

	for _, s := range ui.streams[g.GamePk] {
		streamDisplay.WriteString(s.MediaFeedType + " [" + s.CallLetters + "]\n")
	}

	s = strings.TrimSpace(streamDisplay.String())

	return

}

// GenerateStreamTable shows a list of streams in case there are multiple streams with the same call letters
func (ui *UI) GenerateStreamTable(streams []*Stream) string {

	ts := &strings.Builder{}
	table := tablewriter.NewWriter(ts)
	table.SetAutoMergeCells(true)
	table.SetRowLine(true)
	table.SetColMinWidth(1, 50)

	ts.WriteString("Multiple Games for " + streams[0].MediaFeedType + " [" + streams[0].CallLetters + "]...")
	for _, s := range streams {
		g := ui.schedule.GameMap[s.GamePk]
		table.Append([]string{s.ID, ui.getTeamDisplay(&g, true)})
	}
	table.Render()
	return ts.String()

}

// GetStartStreamlinkDisplay to show details of the selected stream
func (ui *UI) GetStartStreamlinkDisplay(s *Stream) (d string) {
	g := ui.schedule.GameMap[s.GamePk]
	d = "Starting " + s.MediaFeedType + " [" + s.CallLetters + "] stream for " + ui.getTeamDisplay(&g, true) + "..."
	return
}

func (ui *UI) getTeamDisplay(g *Game, singleLine bool) string {

	delim := nl
	if singleLine {
		delim = " vs "
	}
	return g.Teams.Away.Team.Name + " (" + g.Teams.Away.Team.Abbreviation + ")" + delim + g.Teams.Home.Team.Name + " (" + g.Teams.Home.Team.Abbreviation + ")"
}

func (ui *UI) getGameStatusDisplay(g *Game) string {

	sc := g.GameStatus.StatusCode
	sd := &strings.Builder{}

	if sc == "S" || sc == "P" || sc == "PW" {
		// scheduled / pre-game
		t, _ := time.Parse(time.RFC3339, g.GameDate)
		sd.WriteString(timeFormat(&t, false))
		if sc == "PW" {
			// warmup	
			sd.WriteString(nl + g.GameStatus.DetailedState)
		}
	} else if isActiveGame(g.GameStatus.DetailedState) {

		sd.WriteString(g.LineScore.InningState[0:3] + " " + g.LineScore.CurrentInningOrdinal)

		if isDelayedSuspended(g.GameStatus.DetailedState) {
			sd.WriteString(nl + g.GameStatus.DetailedState)
		}

	} else {
		sd.WriteString(g.GameStatus.DetailedState)
	}

	return sd.String()

}

func (ui *UI) getGameScoreDisplay(g *Game) (s string) {

	s = nl

	if hasGameStarted(g.GameStatus.DetailedState) {
		s = strconv.Itoa(g.LineScore.Scoring.Away.Runs) + nl + strconv.Itoa(g.LineScore.Scoring.Home.Runs)
	}

	return
}

// Prompt for user input
func (ui *UI) Prompt() (input string) {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print(">> ")
	input, _ = reader.ReadString('\n')
	input = strings.ToUpper(strings.TrimSpace(input))
	return
}
