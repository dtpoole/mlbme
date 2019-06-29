package main

import (
	"log"
	"os"
	"os/exec"
	"syscall"
)

var proxyCmd *exec.Cmd

func startProxy() {
	if !config.CheckStreams || proxyCmd != nil {
		return
	}

	proxyCmd = exec.Command(proxyPath, "-d", config.Proxy.Domain, "-p", "9876", "-s", config.Proxy.SourceDomains)
	proxyCmd.Env = os.Environ()

	if err := proxyCmd.Start(); err != nil {
		log.Fatal("Unable to start proxy: ", err)
	}
}

func stopProxy() {
	if proxyCmd != nil {
		if err := proxyCmd.Process.Signal(syscall.SIGTERM); err != nil {
			log.Fatal("Unable to stop proxy: ", err)
		}
	}
}
