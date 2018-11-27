package main

import (
	"fmt"
	"sync"
)

type Location struct {
	Latitude  float64 `json:"lat"`
	Longitude float64 `json:"lon"`
	Timestamp int64   `json:"timestamp"`
	Token     string  `json:"-"`
}

var (
	LatestLocation map[Token]Location
	LocationMutex  sync.Mutex
)

func init() {
	LatestLocation = make(map[Token]Location)
}

func (loc Location) String() string {
	return fmt.Sprintf("%s:(%f,%f):[%d]\n", loc.Token, loc.Latitude, loc.Longitude, loc.Timestamp)
}

func StoreLocation(tokenStr string, locs []Location) {
	token := Token(tokenStr)
	if len(locs) > 0 {
		loc := locs[len(locs)-1]
		go func() {
			LocationMutex.Lock()
			defer LocationMutex.Unlock()
			if last, ok := LatestLocation[token]; !ok || last.Timestamp < loc.Timestamp {
				LatestLocation[token] = loc
			}
		}()
	}
	if !LocationDone {
		LocationWait.Add(len(locs))
		go func() {
			for _, loc := range locs {
				loc.Token = tokenStr
				LocationBuffer <- loc
			}
		}()
	}
}

func LatestLocations() []Location {
	LocationMutex.Lock()
	defer LocationMutex.Unlock()
	var result []Location
	for _, loc := range LatestLocation {
		result = append(result, loc)
	}
	return result
}
