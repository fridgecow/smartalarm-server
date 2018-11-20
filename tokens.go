package main

import (
	"bufio"
	"fmt"
	"log"
	"strings"
	"sync"

	"github.com/oliveroneill/exponent-server-sdk-golang/sdk"
)

type Token = expo.ExponentPushToken

var (
	TokenStore map[Token]struct{}
	TokenMutex sync.Mutex

	TokenToId map[Token]string
	IdToToken map[string]Token
	IdMutex   sync.Mutex
)

func tokenInit() {
	TokenStore = make(map[Token]struct{})
	TokenToId = make(map[Token]string)
	IdToToken = make(map[string]Token)

	log.Println("Reading tokens from file...")
	tokenFile := bufio.NewScanner(TokenFile)
	for tokenFile.Scan() {
		TokenStore[Token(tokenFile.Text())] = struct{}{}
	}
	if err := tokenFile.Err(); err != nil {
		log.Fatal("Failed to read token file: ", err)
	}

	log.Println("Reading crsids from file...")
	idFile := bufio.NewScanner(IdFile)
	for idFile.Scan() {
		line := strings.Split(idFile.Text(), ":")
		storeCrsid(Token(line[1]), line[0])
	}
	if err := idFile.Err(); err != nil {
		log.Fatal("Failed to read crsid file: ", err)
	}

	log.Println("Files read")
}

func storeCrsid(token Token, crsid string) {
	if id, seen := TokenToId[token]; seen {
		delete(IdToToken, id)
		delete(TokenToId, token)
	}
	if tok, seen := IdToToken[crsid]; seen {
		delete(TokenToId, tok)
		delete(IdToToken, crsid)
	}
	TokenToId[token] = crsid
	IdToToken[crsid] = token
}

func RegisterToken(tokenStr string) error {
	log.Println("Registering token ", tokenStr)
	token, err := expo.NewExponentPushToken(Token(tokenStr))
	if err != nil {
		return err
	}

	TokenMutex.Lock()
	defer TokenMutex.Unlock()

	TokenStore[token] = struct{}{}
	if _, err := fmt.Fprintf(TokenFile, "%s\n", token); err != nil {
		return Err("Encountered an error writing token to file: %s", err)
	}
	log.Println("Successfully registered token ", tokenStr)
	return nil
}

func RegisterCrsid(tokenStr, crsid string) error {
	log.Println("Registering crsid ", crsid, " to token ", tokenStr)
	if err := RegisterToken(tokenStr); err != nil {
		return err
	}
	token := Token(tokenStr)

	IdMutex.Lock()
	defer IdMutex.Unlock()

	storeCrsid(token, crsid)
	if _, err := fmt.Fprintf(IdFile, "%s:%s\n", crsid, token); err != nil {
		return fmt.Errorf("Encountered an error writing crsid to file: %s", err)
	}
	log.Println("Successfully registered crsid ", crsid, " to token ", tokenStr)
	return nil
}
