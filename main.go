package main

import (
	"encoding/json"
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

type response struct {
	Key string `json:"key"`
	Foo string `json:"foo"`
}

func registerToken(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response{Key: "value", Foo: "bar"})
}

func registerCrsid(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Register worker crsid"))
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
