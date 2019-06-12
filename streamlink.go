package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strconv"
)

var streamlinkCmd *exec.Cmd

func runStreamlink(s Stream) {

	ua := "User-Agent=Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/59.0.3071.115 Safari/537.36"

	streamlinkPath, lookErr := exec.LookPath("streamlink")
	if lookErr != nil {
		panic(lookErr)
	}

	log.Println("Starting streamlink...")

	streamlinkCmd = exec.Command(streamlinkPath, fmt.Sprintf("hlsvariant://%s name_key=bitrate verify=False", s.StreamPlaylist),
		"best", "--http-header", fmt.Sprintf("\"%s\"", ua), "--hls-segment-threads=4", "--https-proxy", "127.0.0.1:"+strconv.Itoa(config.Proxy.Port),
		"--player", config.VLC.Path)

	streamlinkCmd.Env = os.Environ()

	stdout, err := streamlinkCmd.StdoutPipe()
	if err != nil {
		log.Fatal(err)
	}
	if err := streamlinkCmd.Start(); err != nil {
		log.Fatal("Unable to start streamlink: ", err)
	}
	scanner := bufio.NewScanner(stdout)
	scanner.Split(bufio.ScanLines)
	for scanner.Scan() {
		m := scanner.Text()
		log.Println(m)
	}

}
