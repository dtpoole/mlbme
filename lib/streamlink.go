package lib

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"syscall"

	log "github.com/sirupsen/logrus"
)

// Streamlink struct contains execution information streamlink
type Streamlink struct {
	path    string
	cmd     *exec.Cmd
	Running bool
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
		err = errors.New("unable to find streamlink in path")
		return
	}

	log.WithFields(log.Fields{
		"streamlinkPath": s.path,
	}).Debug("NewStreamlink")

	return
}

// Run streamlink
func (s *Streamlink) Run(stream *Stream, http bool) (err error) {

	if s.Running {
		err = errors.New("stream is currently running")
		return
	}

	s.cmd = exec.Command(s.path, fmt.Sprintf("hls://%s name_key=bitrate verify=False", stream.StreamPlaylist),
		"best", "--http-header", fmt.Sprintf("User-Agent=%s", UserAgent), "--hls-segment-threads=4",
		"--https-proxy", "https://127.0.0.1:9876",
		"--player-external-http",
		"--player-external-http-port", "6789",
	)

	s.cmd.Env = os.Environ()

	stdout, err := s.cmd.StdoutPipe()
	if err != nil {
		return
	}
	if err = s.cmd.Start(); err != nil {
		err = errors.New("unable to start streamlink")
		return
	}

	log.WithFields(log.Fields{
		"cmd": strings.Join(s.cmd.Args, " "),
	}).Debug("Started streamlink")

	s.Running = true

	scanner := bufio.NewScanner(stdout)
	scanner.Split(bufio.ScanLines)
	for scanner.Scan() {
		m := scanner.Text()
		log.Debug(m)
		// if 403 assume stream isn't available.
		if match("403 Client Error: Forbidden", m) {
			err = errors.New("Stream is not available")
			s.Stop()
			return
		} else if match("Stream ended", m) {
			fmt.Println("\nStream ended")
			s.Stop()
			return
		}
	}
	return
}

// Stop the streamlink process
func (s *Streamlink) Stop() (err error) {
	if s.Running {
		err = s.cmd.Process.Signal(syscall.SIGTERM)
		s.Running = false
		log.Debug("Stopped streamlink")
	}
	return
}
