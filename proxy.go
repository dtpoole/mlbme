package main

import (
	"log"
	"os"
	"os/exec"
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

	log.Println("Proxy started.")

}

func stopProxy() {
	if proxyCmd != nil {
		if err := proxyCmd.Process.Kill(); err != nil {
			log.Fatal("Unable to stop proxy: ", err)
		}
		log.Println("Proxy stopped.")
	}
}
