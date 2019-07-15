package lib

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"os/exec"
	"syscall"

	"github.com/pkg/errors"
)

// Streamlink struct contains execution information streamlink/vlc
type Streamlink struct {
	path    string
	cmd     *exec.Cmd
	running bool
	vlcPath string
}

// NewStreamlink creates initialize the Streamlink struct
func NewStreamlink() (s Streamlink, err error) {

	streamlinkPaths := []string{"streamlink", "/usr/local/bin/streamlink"}
	for _, path := range streamlinkPaths {
		if s.path, err = exec.LookPath(path); err == nil {
			break
		}
	}

	if s.path == "" {
		err = errors.New(" Unable to find streamlink in path")
		return
	}

	vlcPaths := []string{"cvlc", "vlc", "/Applications/VLC.app/Contents/MacOS/VLC", "~/Applications/VLC.app/Contents/MacOS/VLC"}
	for _, path := range vlcPaths {
		if s.vlcPath, err = exec.LookPath(path); err == nil {
			break
		}
	}

	if s.vlcPath == "" {
		err = errors.New(" Unable to find VLC in path")
		return
	}

	return
}

// Run streamlink
func (s *Streamlink) Run(stream *Stream, http bool) (err error) {

	if http || match("cvlc", s.vlcPath) {
		log.Println("HTTP streaming enabled.")
		s.vlcPath = s.vlcPath + " --sout '#standard{access=http,mux=ts,dst=:6789}'"
	}

	s.cmd = exec.Command(s.path, fmt.Sprintf("hls://%s name_key=bitrate verify=False", stream.StreamPlaylist),
		"best", "--http-header", fmt.Sprintf("User-Agent=%s", UserAgent), "--hls-segment-threads=4",
		"--https-proxy", "127.0.0.1:9876",
		"--player", s.vlcPath)

	s.cmd.Env = os.Environ()

	stdout, err := s.cmd.StdoutPipe()
	if err != nil {
		log.Fatal(err)
	}
	if err := s.cmd.Start(); err != nil {
		log.Fatal("Unable to start streamlink: ", err)
	}

	s.running = true

	scanner := bufio.NewScanner(stdout)
	scanner.Split(bufio.ScanLines)
	for scanner.Scan() {
		m := scanner.Text()
		log.Println(m)
		// if 403 assume stream isn't available.
		if match("403 Client Error: Forbidden", m) {
			log.Println("Stream is not available.")
			s.Stop()
		}
	}
	return
}

// Stop the streamlink process
func (s *Streamlink) Stop() (err error) {
	if s.running {
		err = s.cmd.Process.Signal(syscall.SIGTERM)
		s.running = false
	}
	return
}
