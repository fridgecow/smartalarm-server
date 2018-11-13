package main

import (
	"io"
	"log"
	"os"
	"os/signal"
	"time"
)

var (
	LogFile   *os.File
	TokenFile *os.File
)

func init() {
	var err error

	LogFile, err = os.OpenFile(time.Now().String()+".log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		panic(err)
	}
	log.SetOutput(io.MultiWriter(os.Stdout, LogFile))

	TokenFile, err = os.OpenFile("tokens", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		panic(err)
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		for sig := range c {
			log.Printf("Received OS interrupt: %s. Committing changes to disk and quitting.", sig)
			LogFile.Close()
			TokenFile.Close()
			// TODO: Ensure all data is pushed to disk
			os.Exit(1)
		}
	}()
}
