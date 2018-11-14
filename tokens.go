package main

import (
	"fmt"
)

var (
	TokenStore map[string]struct{}

	TokenToId map[string]string
	IdToToken map[string]string
)

func tokenInit() {
	TokenStore = make(map[string]struct{})
	TokenToId = make(map[string]string)
	IdToToken = make(map[string]string)
	// TODO: Read in from file
}

func RegisterToken(token string) error {
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

	if id, seen := TokenToId[token]; seen {
		delete(IdToToken, id)
	}
	if tok, seen := IdToToken[crsid]; seen {
		delete(TokenToId, tok)
	}
	TokenToId[token] = crsid
	IdToToken[crsid] = token

	if _, err := fmt.Fprintf(IdFile, "%s:%s\n", crsid, token); err != nil {
		return fmt.Errorf("Encountered an error writing crsid to file: %s", err)
	}

	return nil
}
