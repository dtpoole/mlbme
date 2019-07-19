package main

import (
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/alexflint/go-arg"
	"github.com/dtpoole/mlbme/lib"
	log "github.com/sirupsen/logrus"
)

var (
	config      *lib.Config
	schedule    lib.Schedule
	proxy       lib.Proxy
	streamlink  lib.Streamlink
	gamestreams lib.GameStreams
	ui          lib.UI
	err         error
	version     string
)

type args struct {
	Config string `arg:"-c" help:"JSON configuration"`
	HTTP   bool   `help:"use HTTP streaming instead of playing locally"`
	Team   string `arg:"-t" help:"filter on team by abbreviation"`
	Stream string `arg:"-s" help:"call letter of stream to start"`
	Debug  bool   `arg:"-d" help:"enable debug logging"`
}

// consts
const (
	RefreshRate = 5 * time.Minute
)

func init() {
	// handle ctrl-c (sigterm)
	stCh := make(chan os.Signal)
	signal.Notify(stCh, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-stCh
		streamlink.Stop()
		proxy.Stop()
		os.Exit(1)
	}()
}

func refresh(periodic bool) {
	r := func() {
		schedule, err = lib.GetMLBSchedule(config.StatsURL)
		if err != nil {
			exit(err)
		}

		if config.CheckStreams {
			gamestreams.GetAvailableStreams()
		}
	}

	if !periodic {
		r()
	} else {
		ticker := time.NewTicker(RefreshRate)
		for range ticker.C {
			r()
			fmt.Print(ui.GenerateScoreboard())
		}
	}
}

func startStream(streamID string, http bool) {
	var strs []*lib.Stream

	for _, gs := range gamestreams.Streams {
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
		fmt.Println(ui.GetStartStreamlinkDisplay(strs[0]))
		go streamlink.Run(strs[0], http)
	default:
		fmt.Println(ui.GenerateStreamTable(strs))
	}
}

func exit(err error) {
	code := 0
	if err != nil {
		code = 1
		log.Println(err)
	}
	streamlink.Stop()
	proxy.Stop()
	os.Exit(code)
}

func (args) Version() string {
	return "mlbme " + version
}

func main() {

	var args args
	args.Config = "config.json"
	arg.MustParse(&args)

	log.SetOutput(os.Stdout)
	log.SetFormatter(&log.TextFormatter{
		DisableColors: true,
		FullTimestamp: true,
	})

	if args.Debug {
		log.SetLevel(log.DebugLevel)
	}

	log.Debug("Debug logging enabled")

	config, err = lib.LoadConfig(args.Config)
	if err != nil {
		exit(err)
	}

	if config.CheckStreams {

		proxy, err = lib.NewProxy(config)
		if err != nil {
			exit(err)
		}

		streamlink, err = lib.NewStreamlink()
		if err != nil {
			exit(err)
		}

		gamestreams = lib.NewGameStreams(config, &schedule)

		proxy.Run()

	}

	refresh(false)

	ui = lib.NewUI(config, &schedule, gamestreams.Streams, args.Team)

	fmt.Print(ui.GenerateScoreboard())

	// setup background refresh
	go refresh(true)

	if args.Stream != "" {
		startStream(strings.ToUpper(args.Stream), args.HTTP)
	}

	for {
		input := ui.Prompt()
		if input == "Q" {
			exit(nil)
		} else if input == "R" || input == "" {
			fmt.Print(ui.GenerateScoreboard())
		} else if input == "H" {
			fmt.Println("[call letters] = play stream\nr = refresh\nq = quit")
		} else {
			startStream(input, args.HTTP)
		}
	}
}
