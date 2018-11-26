package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

func test(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("<html><body><h1>Test Page</h1><h2>Tokens</h2><ul>"))
	for tok := range TokenStore {
		w.Write([]byte("<li>" + string(tok) + "</li>"))
	}
	w.Write([]byte("</ul><h2>Tokens -> Crsids</h2><ul>"))
	for tok, id := range TokenToId {
		w.Write([]byte("<li>" + string(tok) + " -> " + id + "</li>"))
	}
	w.Write([]byte("</ul><h2>Crsids -> Tokens</h2><ul>"))
	for id, tok := range IdToToken {
		w.Write([]byte("<li>" + id + " -> " + string(tok) + "</li>"))
	}
	w.Write([]byte("</ul></body></html>"))
}

type Response struct {
	Ok     bool     `json:"ok,omitempty"`
	Errors []string `json:"errors,omitempty"`
}

var Err = fmt.Errorf

func handler(f func(*http.Request) []error) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		errs := f(r)
		var errStr []string
		for _, err := range errs {
			errStr = append(errStr, fmt.Sprintf("%s", err))
		}
		json.NewEncoder(w).Encode(Response{Ok: errs == nil, Errors: errStr})
	}
}

func registerToken(r *http.Request) []error {
	decoder := json.NewDecoder(r.Body)
	data := make(map[string]string)
	if err := decoder.Decode(&data); err != nil {
		buf := new(bytes.Buffer)
		buf.ReadFrom(r.Body)
		s := buf.String()
		return []error{Err("Error decoding json request. Error: [%s]. Request: [%s].", err, s)}
	}
	token, ok := data["token"]
	if !ok {
		return []error{Err("Key 'token' not found in request")}
	}
	if err := RegisterToken(token); err != nil {
		return []error{err}
	}
	return nil
}

func registerCrsid(r *http.Request) []error {
	decoder := json.NewDecoder(r.Body)
	data := make(map[string]string)
	if err := decoder.Decode(&data); err != nil {
		buf := new(bytes.Buffer)
		buf.ReadFrom(r.Body)
		s := buf.String()
		return []error{Err("Error decoding json request. Error: [%s]. Request: [%s].", err, s)}
	}
	token, ok := data["token"]
	if !ok {
		return []error{Err("Key 'token' not found in request")}
	}
	crsid, ok := data["crsid"]
	if !ok {
		return []error{Err("Key 'crsid' not found in request")}
	}
	if err := RegisterCrsid(token, crsid); err != nil {
		return []error{err}
	}
	return nil
}

type PushRequest struct {
	Tokens []string          `json:"tokens"`
	Crsids []string          `json:"crsids"`
	Title  string            `json:"title"`
	Body   string            `json:"body"`
	Data   map[string]string `json:"data"`
}

func pushAll(r *http.Request) []error {
	decoder := json.NewDecoder(r.Body)
	var request PushRequest
	if err := decoder.Decode(&request); err != nil {
		buf := new(bytes.Buffer)
		buf.ReadFrom(r.Body)
		s := buf.String()
		return []error{Err("Error decoding json request. Error: [%s]. Request: [%s].", err, s)}
	}
	var tokens []Token
	for token := range TokenStore {
		tokens = append(tokens, token)
	}
	return pushToTokens(tokens, request.Title, request.Body, request.Data)
}

func pushTokens(r *http.Request) []error {
	decoder := json.NewDecoder(r.Body)
	var request PushRequest
	if err := decoder.Decode(&request); err != nil {
		buf := new(bytes.Buffer)
		buf.ReadFrom(r.Body)
		s := buf.String()
		return []error{Err("Error decoding json request. Error: [%s]. Request: [%s].", err, s)}
	}
	var tokens []Token
	for _, token := range request.Tokens {
		tokens = append(tokens, Token(token))
	}
	return pushToTokens(tokens, request.Title, request.Body, request.Data)
}

func pushCrsids(r *http.Request) []error {
	decoder := json.NewDecoder(r.Body)
	var request PushRequest
	if err := decoder.Decode(&request); err != nil {
		buf := new(bytes.Buffer)
		buf.ReadFrom(r.Body)
		s := buf.String()
		return []error{Err("Error decoding json request. Error: [%s]. Request: [%s].", err, s)}
	}
	return pushToCrsids(request.Crsids, request.Title, request.Body, request.Data)
}

type StoreLocationRequest struct {
	Token     string     `json:"token"`
	Crsid     string     `json:"string"`
	Locations []Location `json:"locations"`
}

func storeLocations(r *http.Request) []error {
	decoder := json.NewDecoder(r.Body)
	var request StoreLocationRequest
	if err := decoder.Decode(&request); err != nil {
		buf := new(bytes.Buffer)
		buf.ReadFrom(r.Body)
		s := buf.String()
		return []error{Err("Error decoding json request. Error: [%s]. Request: [%s].", err, s)}
	}
	StoreLocation(request.Token, request.Crsid, request.Locations)
	return nil
}

func serveLocations(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "log/locations")
}

func main() {
	log.Println("Initialisation complete")
	http.HandleFunc("/register/token", handler(registerToken))
	http.HandleFunc("/register/crsid", handler(registerCrsid))
	http.HandleFunc("/push/all", handler(pushAll))
	http.HandleFunc("/push/tokens", handler(pushTokens))
	http.HandleFunc("/push/crsids", handler(pushCrsids))
	http.HandleFunc("/locations/store", handler(storeLocations))
	http.HandleFunc("/locations/get", serveLocations)
	http.HandleFunc("/test", test)

	log.Println("Server now running")
	log.Fatal(http.ListenAndServe(":6662", nil))
}
