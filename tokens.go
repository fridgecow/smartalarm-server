package main

import (
	"bufio"
	"fmt"
	"log"
	"strings"
	"sync"
)

var (
	TokenStore map[string]struct{}
	TokenMutex sync.Mutex

	TokenToId map[string]string
	IdToToken map[string]string
	IdMutex   sync.Mutex
)

func tokenInit() {
	TokenStore = make(map[string]struct{})
	TokenToId = make(map[string]string)
	IdToToken = make(map[string]string)

	tokenFile := bufio.NewScanner(TokenFile)
	for tokenFile.Scan() {
		TokenStore[tokenFile.Text()] = struct{}{}
	}
	if err := tokenFile.Err(); err != nil {
		log.Fatal("Failed to read token file: ", err)
	}

	idFile := bufio.NewScanner(IdFile)
	for idFile.Scan() {
		line := strings.Split(idFile.Text(), ":")
		storeCrsid(line[1], line[0])
	}
	if err := idFile.Err(); err != nil {
		log.Fatal("Failed to read crsid file: ", err)
	}
}

func storeCrsid(token, crsid string) {
	if id, seen := TokenToId[token]; seen {
		delete(IdToToken, id)
	}
	if tok, seen := IdToToken[crsid]; seen {
		delete(TokenToId, tok)
	}
	TokenToId[token] = crsid
	IdToToken[crsid] = token
}

func RegisterToken(token string) error {
	TokenMutex.Lock()
	defer TokenMutex.Unlock()

	TokenStore[token] = struct{}{}
	if _, err := fmt.Fprintf(TokenFile, "%s\n", token); err != nil {
		return fmt.Errorf("Encountered an error writing token to file: %s", err)
	}
	return nil
}

func RegisterCrsid(token, crsid string) error {
	if err := RegisterToken(token); err != nil {
		return err
	}

	IdMutex.Lock()
	defer IdMutex.Unlock()

	storeCrsid(token, crsid)
	if _, err := fmt.Fprintf(IdFile, "%s:%s\n", crsid, token); err != nil {
		return fmt.Errorf("Encountered an error writing crsid to file: %s", err)
	}
	return nil
}
