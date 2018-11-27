package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"sync"
	"time"
)

var (
	LogFile      *os.File
	TokenFile    *os.File
	IdFile       *os.File
	LocationFile *os.File

	LocationBuffer chan Location
	LocationWait   sync.WaitGroup
	LocationDone   bool
)

func init() {
	log.Println("Initialising server")
	var err error

	LogFile, err = os.OpenFile(time.Now().Format("log/2006-01-02.15-04-05.log"),
		os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0440)
	if err != nil {
		log.Fatal(err)
	}
	log.SetOutput(io.MultiWriter(os.Stdout, LogFile))
	log.Println("Logging to file")

	TokenFile, err = os.OpenFile("log/tokens", os.O_APPEND|os.O_CREATE|os.O_RDWR, 0460)
	if err != nil {
		log.Fatal(err)
	}
	IdFile, err = os.OpenFile("log/crsids", os.O_APPEND|os.O_CREATE|os.O_RDWR, 0460)
	if err != nil {
		log.Fatal(err)
	}

	LocationFile, err = os.OpenFile("log/locations", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0460)
	if err != nil {
		log.Fatal(err)
	}
	LocationBuffer = make(chan Location, 512)
	go func() {
		for location := range LocationBuffer {
			if _, err := fmt.Fprint(LocationFile, location); err != nil {
				// do anything?
			}
			LocationWait.Done()
		}
	}()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		for sig := range c {
			log.Printf("Received OS interrupt: %s. Cleaning up, ensuring changes are committed to disk, and quitting.", sig)

			LogFile.Close()
			TokenFile.Close()
			IdFile.Close()

			LocationDone = true
			LocationWait.Wait()
			LocationFile.Close()

			os.Exit(1)
		}
	}()
	tokenInit()
}
