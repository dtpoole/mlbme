package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/dtpoole/mlbme/lib"
)

var (
	config     *lib.Config
	schedule   lib.Schedule
	streams    map[int]map[string]*lib.Stream
	proxy      lib.Proxy
	streamlink lib.Streamlink
	ui         lib.UI
	err        error
	version    string
)

var (
	configFlag  = flag.String("config", "config.json", "JSON configuration")
	httpFlag    = flag.Bool("http", false, "Tell VLC to use HTTP streaming instead of playing locally")
	teamFlag    = flag.String("team", "", "Filter on team by abbreviation")
	streamFlag  = flag.String("stream", "", "Call letter of stream to start")
	versionFlag = flag.Bool("v", false, "Display version")
)

// consts
const (
	RefreshRate = 3 * time.Minute
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
			streams = lib.GetAvailableStreams(config, &schedule)
			// if err != nil {
			// 	exit(err)
			// }
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
	var strs []*lib.Stream

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

func main() {

	flag.Parse()

	if *versionFlag {
		fmt.Println(version)
		os.Exit(0)
	}

	config, err = lib.LoadConfig(*configFlag)
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

		proxy.Run()

	}

	refresh(false)

	ui = lib.NewUI(config, &schedule, streams, *teamFlag)

	fmt.Print(ui.GenerateScoreboard())

	// setup background refresh
	go refresh(true)

	if *streamFlag != "" {
		startStream(strings.ToUpper(*streamFlag), *httpFlag)
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
			startStream(input, *httpFlag)
		}
	}
}
