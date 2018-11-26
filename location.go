package main

import (
	"log"
	"sync"
)

type Location struct {
	Latitude  float64 `json:"lat"`
	Longitude float64 `json:"lon"`
	TimeStamp int64   `json:"timestamp"`
	Token     string  `json:"-"`
}

func (loc Location) String() string {
	return fmt.Sprintf(LocationFile, "%s:(%f,%f):[%s]\n", loc.Token, loc.Latitude, loc.Longitude, loc.TimeStamp)
}

func StoreLocation(tokenStr string, crsid string, locs []Location) {
	go func() {
		for _, loc := range locs {
			loc.Token = tokenStr
			LocationBuffer <- loc
		}
	}()
	// store crsid?
}
