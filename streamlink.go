package main

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"syscall"
)

func runStreamlink(s Stream) {

	var args []string

	// TODO: check on startup and set config
	streamlinkPath, lookErr := exec.LookPath("streamlink")
	if lookErr != nil {
		panic(lookErr)
	}

	ua := "User-Agent=Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/59.0.3071.115 Safari/537.36"

	args = append(args, streamlinkPath)

	args = append(args, fmt.Sprintf("hlsvariant://%s name_key=bitrate verify=False", s.StreamPlaylist), "best")

	args = append(args, "--http-header", fmt.Sprintf("\"%s\"", ua))
	args = append(args, "--hls-segment-threads=4")
	args = append(args, "--https-proxy", "127.0.0.1:"+strconv.Itoa(config.Proxy.Port))

	// TODO: add config option for http stream
	args = append(args, "--player", config.VLC.Path)
	//args = append(args, "--player", config.VLC.Path+" --sout '#standard{access=http,mux=ts,dst=:6789}'")

	fmt.Println("Starting streamlink...")
	//fmt.Println(strings.Join(args, " \\\n"))

	env := os.Environ()
	execErr := syscall.Exec(streamlinkPath, args, env)
	if execErr != nil {
		panic(execErr)
	}

}
