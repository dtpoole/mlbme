package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"os/exec"
	"time"
)

var streamlinkCmd *exec.Cmd

func runStreamlink(s Stream, http bool) {

	vlc := vlcPath

	if http || match("cvlc", vlcPath) {
		log.Println("HTTP streaming enabled.")
		vlc = vlc + " --sout '#standard{access=http,mux=ts,dst=:6789}'"
	}

	ua := "User-Agent=Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/59.0.3071.115 Safari/537.36"

	log.Println("Starting streamlink...")

	streamlinkCmd = exec.Command(streamlinkPath, fmt.Sprintf("hlsvariant://%s name_key=bitrate verify=False", s.StreamPlaylist),
		"best", "--http-header", fmt.Sprintf("\"%s\"", ua), "--hls-segment-threads=4", "--https-proxy", "127.0.0.1:9876",
		"--player", vlc)

	streamlinkCmd.Env = os.Environ()

	stdout, err := streamlinkCmd.StdoutPipe()
	if err != nil {
		log.Fatal(err)
	}
	if err := streamlinkCmd.Start(); err != nil {
		log.Fatal("Unable to start streamlink: ", err)
	}

	currentlyPlayingID = s.ID

	scanner := bufio.NewScanner(stdout)
	scanner.Split(bufio.ScanLines)
	for scanner.Scan() {
		m := scanner.Text()

		// check for error opening commerical break stream. if it occurs, sleep for 60 seconds and restart stream.
		// TODO: check to see if streamlink can handle this
		if match("Unable to open URL", m) {
			log.Println("ERROR: Unable to open URL. Will reopen stream in 60 secs")
			streamlinkCmd.Process.Kill()
			time.Sleep(60 * time.Second)
			runStreamlink(s, http)
		}
		log.Println(m)
	}

}
