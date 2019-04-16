package main

import (
	"io"
	"log"
	"os"
	"os/signal"
	"time"
)

var LogFile *os.File

func init() {
	log.Println("Initialising server")
	var err error

	fileName := time.Now().Format("log/2006-01-02.15-04-05.log")
	LogFile, err = os.OpenFile(fileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0440)
	if err != nil {
		log.Fatal(err)
	}
	log.SetOutput(io.MultiWriter(os.Stdout, LogFile))
	log.Println("Created log file", fileName)

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		for sig := range c {
			log.Printf("Received OS interrupt: %s. Cleaning up, ensuring changes are committed to disk, and quitting.", sig)

			LogFile.Close()
			os.Exit(1)
		}
	}()
}
