package main

import (
	"log"
	"os"
	"os/exec"
	"strconv"
)

var proxyCmd *exec.Cmd

func startProxy() {

	if proxyCmd != nil {
		return
	}

	proxyCmd = exec.Command(proxyPath, "-d", config.Proxy.Domain, "-p", strconv.Itoa(config.Proxy.Port), "-s", config.Proxy.SourceDomains)
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
