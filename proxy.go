package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strconv"
	"time"
)

var proxyCmd *exec.Cmd

func startProxy() {

	fmt.Println("Starting proxy...")

	proxyCmd = exec.Command(config.Proxy.Path, "-d", config.Proxy.Domain, "-p", strconv.Itoa(config.Proxy.Port), "-s", config.Proxy.SourceDomains)
	//proxyCmd = exec.Command(proxyPath, "-debug", "-d", config.Proxy.Domain, "-p", config.Proxy.Port, "-s", config.Proxy.SourceDomains)
	proxyCmd.Env = os.Environ()

	if err := proxyCmd.Start(); err != nil {
		log.Fatal("Unable to start proxy: ", err)
	}

	// sleep to ensure started before stream
	time.Sleep(2 * time.Second)

}

func stopProxy() {
	fmt.Println("Stopping proxy...")
	if err := proxyCmd.Process.Kill(); err != nil {
		log.Fatal("Unable to stop proxy process: ", err)
	}
}
