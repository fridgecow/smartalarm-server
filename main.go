package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
)

func sayHello(w http.ResponseWriter, r *http.Request) {
	message := r.URL.Path
	message = strings.TrimPrefix(message, "/")
	message = "Hello " + message

	w.Write([]byte(message))
}

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
	Success string `json:"success,omitempty"`
	Error   string `json:"error,omitempty"`
}

func str(x interface{}) string {
	if x == nil {
		return ""
	}
	return fmt.Sprintf("%s", x)
}

func Return(w io.Writer, success, err interface{}) {
	json.NewEncoder(w).Encode(Response{Success: str(success), Error: str(err)})
}

func registerToken(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	decoder := json.NewDecoder(r.Body)
	data := make(map[string]string)
	if err := decoder.Decode(&data); err != nil {
		buf := new(bytes.Buffer)
		buf.ReadFrom(r.Body)
		s := buf.String()
		Return(w, nil, fmt.Sprintf("Error decoding json request. Error: [%s]. Request: [%s].", err, s))
		return
	}
	token, ok := data["token"]
	if !ok {
		Return(w, nil, "Key 'token' not found in request")
		return
	}
	if err := RegisterToken(token); err != nil {
		Return(w, nil, err)
		return
	}
	Return(w, token, nil)
}

func registerCrsid(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	decoder := json.NewDecoder(r.Body)
	data := make(map[string]string)
	if err := decoder.Decode(&data); err != nil {
		buf := new(bytes.Buffer)
		buf.ReadFrom(r.Body)
		s := buf.String()
		Return(w, nil, fmt.Sprintf("Error decoding json request. Error: [%s]. Request: [%s].", err, s))
		return
	}
	token, ok := data["token"]
	if !ok {
		Return(w, nil, "Key 'token' not found in request")
		return
	}
	crsid, ok := data["crsid"]
	if !ok {
		Return(w, nil, "Key 'crsid' not found in request")
		return
	}
	if err := RegisterCrsid(token, crsid); err != nil {
		Return(w, nil, err)
		return
	}
	Return(w, fmt.Sprintf("%s:%s", crsid, token), nil)
}

type PushRequest struct {
	Tokens []string          `json:"tokens"`
	Crsids []string          `json:"crsids"`
	Title  string            `json:"title"`
	Body   string            `json:"body"`
	Data   map[string]string `json:"data"`
}

func pushAll(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	decoder := json.NewDecoder(r.Body)
	var request PushRequest
	if err := decoder.Decode(&request); err != nil {
		buf := new(bytes.Buffer)
		buf.ReadFrom(r.Body)
		s := buf.String()
		Return(w, nil, fmt.Sprintf("Error decoding json request. Error: [%s]. Request: [%s].", err, s))
		return
	}
	var tokens []Token
	for token := range TokenStore {
		tokens = append(tokens, token)
	}
	json.NewEncoder(w).Encode(pushToTokens(tokens, request.Title, request.Body, request.Data))
}

func pushTokens(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	decoder := json.NewDecoder(r.Body)
	var request PushRequest
	if err := decoder.Decode(&request); err != nil {
		buf := new(bytes.Buffer)
		buf.ReadFrom(r.Body)
		s := buf.String()
		Return(w, nil, fmt.Sprintf("Error decoding json request. Error: [%s]. Request: [%s].", err, s))
		return
	}
	var tokens []Token
	for _, token := range request.Tokens {
		tokens = append(tokens, Token(token))
	}
	json.NewEncoder(w).Encode(pushToTokens(tokens, request.Title, request.Body, request.Data))
}

func pushCrsids(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	decoder := json.NewDecoder(r.Body)
	var request PushRequest
	if err := decoder.Decode(&request); err != nil {
		buf := new(bytes.Buffer)
		buf.ReadFrom(r.Body)
		s := buf.String()
		Return(w, nil, fmt.Sprintf("Error decoding json request. Error: [%s]. Request: [%s].", err, s))
		return
	}
	json.NewEncoder(w).Encode(pushToCrsids(request.Crsids, request.Title, request.Body, request.Data))
}

func main() {
	http.HandleFunc("/register/token", registerToken)
	http.HandleFunc("/register/crsid", registerCrsid)
	http.HandleFunc("/push/all", pushAll)
	http.HandleFunc("/push/tokens", pushTokens)
	http.HandleFunc("/push/crsids", pushCrsids)
	http.HandleFunc("/test", test)
	http.HandleFunc("/", sayHello)

	log.Fatal(http.ListenAndServe(":6662", nil))
}
