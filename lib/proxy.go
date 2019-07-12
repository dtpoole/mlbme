package lib

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"

	"github.com/pkg/errors"
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

// NewProxy creates initialize the Proxy struct
func NewProxy(c *Config) (p Proxy, err error) {

	proxyPaths := []string{"go-mlbam-proxy", "go-mlbam-proxy/go-mlbam-proxy", "/usr/local/bin/go-mlbam-proxy"}
	for _, path := range proxyPaths {
		if p.path, err = exec.LookPath(path); err == nil {
			break
		}
	}

	if p.path == "" {
		err = errors.New("Unable to find go-mlbam-proxy in path")
		return
	}

	p.domain = c.Proxy.Domain
	p.sourceDomains = c.Proxy.SourceDomains
	p.port = "9876"

	p.cmd = exec.Command(p.path, "-d", p.domain, "-p", p.port, "-s", p.sourceDomains)
	p.cmd.Env = os.Environ()

	return
}

// Run the proxy
func (p *Proxy) Run() (err error) {
	if err = p.cmd.Start(); err != nil {
		p.running = false
	} else {
		p.running = true
	}
	return
}

// Stop the proxy
func (p *Proxy) Stop() (err error) {
	if p.running {
		err = p.cmd.Process.Signal(syscall.SIGTERM)
		p.running = false
	}
	return
}

func (p *Proxy) Info() {
	fmt.Println(p.path, "\n", p.running)
}
