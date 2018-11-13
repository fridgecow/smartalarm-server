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
	http.HandleFunc("/", sayHello)

	log.Fatal(http.ListenAndServe(":6662", nil))
}
