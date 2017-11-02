package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"runtime"
	"wireguard"
)

func printUsage() {
	fmt.Printf("usage:\n")
	fmt.Printf("%s [-f/--foreground] INTERFACE-NAME\n", os.Args[0])
}

func main() {

	// parse arguments

	var foreground bool
	var interfaceName string
	if len(os.Args) < 2 || len(os.Args) > 3 {
		printUsage()
		return
	}

	switch os.Args[1] {

	case "-f", "--foreground":
		foreground = true
		if len(os.Args) != 3 {
			printUsage()
			return
		}
		interfaceName = os.Args[2]

	default:
		foreground = false
		if len(os.Args) != 2 {
			printUsage()
			return
		}
		interfaceName = os.Args[1]
	}

	// daemonize the process

	if !foreground {
		err := wireguard.Daemonize()
		if err != nil {
			log.Println("Failed to daemonize:", err)
		}
		return
	}

	// increase number of go workers (for Go <1.5)

	runtime.GOMAXPROCS(runtime.NumCPU())

	// open TUN device

	tun, err := wireguard.CreateTUN(interfaceName)
	if err != nil {
		log.Println("Failed to create tun device:", err)
		return
	}

	// get log level (default: info)

	logLevel := func() int {
		switch os.Getenv("LOG_LEVEL") {
		case "debug":
			return wireguard.LogLevelDebug
		case "info":
			return wireguard.LogLevelInfo
		case "error":
			return wireguard.LogLevelError
		}
		return wireguard.LogLevelInfo
	}()

	// create wireguard device

	device := wireguard.NewDevice(tun, logLevel)

	logInfo := device.Log.Info
	logError := device.Log.Error
	logInfo.Println("Starting device")

	// start configuration lister

	uapi, err := wireguard.NewUAPIListener(interfaceName)
	if err != nil {
		logError.Fatal("UAPI listen error:", err)
	}

	errs := make(chan error)
	term := make(chan os.Signal)
	wait := device.WaitChannel()

	go func() {
		for {
			conn, err := uapi.Accept()
			if err != nil {
				errs <- err
				return
			}
			go wireguard.IpcHandle(device, conn)
		}
	}()

	logInfo.Println("UAPI listener started")

	// wait for program to terminate

	signal.Notify(term, os.Kill)
	signal.Notify(term, os.Interrupt)

	select {
	case <-wait:
	case <-term:
	case <-errs:
	}

	// clean up UAPI bind

	uapi.Close()

	logInfo.Println("Closing")
}
