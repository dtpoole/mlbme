package lib

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"syscall"

	log "github.com/sirupsen/logrus"
)

// Proxy struct contains execution information for the proxy
type Proxy struct {
	path          string
	domain        string
	sourceDomains string
	port          string
	running       bool
	cmd           *exec.Cmd
}

// NewProxy initializes the Proxy struct
func NewProxy(c *Config) (p Proxy, err error) {

	proxyPaths := []string{"go-mlbam-proxy", "go-mlbam-proxy/go-mlbam-proxy", "/usr/local/bin/go-mlbam-proxy"}
	for _, path := range proxyPaths {
		if p.path, err = exec.LookPath(path); err == nil {
			break
		}
	}

	if p.path == "" {
		err = errors.New("unable to find go-mlbam-proxy in path")
		return
	}

	p.domain = c.Proxy.Domain
	p.sourceDomains = c.Proxy.SourceDomains
	p.port = "9876"

	p.cmd = exec.Command(p.path, "-d", p.domain, "-p", p.port, "-s", p.sourceDomains)
	p.cmd.Env = os.Environ()

	log.WithFields(log.Fields{
		"path":          p.path,
		"domain":        p.domain,
		"sourceDomains": p.sourceDomains,
		"port":          p.port,
	}).Debug("NewProxy")

	return
}

// Run the proxy
func (p *Proxy) Run() (err error) {
	if err = p.cmd.Start(); err != nil {
		p.running = false
	} else {
		p.running = true
		log.WithFields(log.Fields{
			"cmd": strings.Join(p.cmd.Args, " "),
		}).Debug("Started proxy")
	}
	return
}

// Stop the proxy
func (p *Proxy) Stop() (err error) {
	if p.running {
		err = p.cmd.Process.Signal(syscall.SIGTERM)
		p.running = false
		log.Debug("Stopped proxy")
	}
	return
}

// Status of the proxy
func (p *Proxy) Status() {
	fmt.Println(p.path, "\n", p.running)
}
