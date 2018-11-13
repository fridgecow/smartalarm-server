package main

import ()

var (
	TokenStore map[string]struct{}

	TokenToId map[string]string
	IdToToken map[string]string
)

func init() {
	TokenStore = make(map[string]struct{})
	TokenToId = make(map[string]string)
	IdToToken = make(map[string]string)
	// TODO: Read in from file
}

func RegisterToken(token string) error {
	TokenStore[token] = struct{}{}
	return nil
	// TODO: Write to file
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

	return nil
	// TODO: Write to file
}
