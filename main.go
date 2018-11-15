package main

import (
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
	for k := range TokenStore {
		w.Write([]byte("<li>" + k + "</li>"))
	}
	w.Write([]byte("</ul><h2>Tokens -> Crsids</h2><ul>"))
	for tok, id := range TokenToId {
		w.Write([]byte("<li>" + tok + " -> " + id + "</li>"))
	}
	w.Write([]byte("</ul><h2>Crsids -> Tokens</h2><ul>"))
	for id, tok := range IdToToken {
		w.Write([]byte("<li>" + id + " -> " + tok + "</li>"))
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
		Return(w, nil, err)
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
		Return(w, nil, err)
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

func pushTokens(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Push notification to tokens"))
}

func pushCrsids(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Push notification to worker crsids"))
}

func main() {
	http.HandleFunc("/register/token", registerToken)
	http.HandleFunc("/register/crsid", registerCrsid)
	http.HandleFunc("/push/tokens", pushTokens)
	http.HandleFunc("/push/crsids", pushCrsids)
	http.HandleFunc("/test", test)
	http.HandleFunc("/", sayHello)

	log.Fatal(http.ListenAndServe(":6662", nil))
}
