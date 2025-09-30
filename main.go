package main

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"html/template"
	"log"
	"math"
	"net/http"
	"time"
)

var templates *template.Template
var db *sql.DB

func main() {
	var err error

	// Database Connection
	db, err = getDatabase()

	if err != nil {
		log.Fatalf("Error on initializing database connection: %s", err.Error())
	}

	// Load templates
	templates, err = template.New("").Funcs(template.FuncMap{
		"dict": func(values ...interface{}) (map[string]interface{}, error) {
			if len(values)%2 != 0 {
				return nil, Err("invalid dict call")
			}
			dict := make(map[string]interface{}, len(values)/2)
			for i := 0; i < len(values); i += 2 {
				key, ok := values[i].(string)
				if !ok {
					return nil, Err("dict keys must be strings")
				}
				dict[key] = values[i+1]
			}
			return dict, nil
		},
		"percent": func(value float64) string {
			return fmt.Sprintf("%.1f", value*100)
		},
		"time": func(value time.Time) string {
			return value.Format("15:04")
		},
		"duration": func(value time.Duration) string {
			d := value.Round(time.Minute)
			h := d / time.Hour
			d -= h * time.Hour
			m := d / time.Minute
			return fmt.Sprintf("%d hours %d minutes", h, m)
		},
		"multDuration": func(a float64, b time.Duration) string {
			duration := fmt.Sprintf("%v", time.Duration(math.Round(a*float64(b))))
			return duration[:len(duration)-2] // Remove "0s" from end
		},
	}).ParseGlob("templates/*.tmpl")
	if err != nil {
		log.Fatalf("Could not parse templates: %s", err.Error())
	}

	log.Println("Initialisation complete")

	r := getRouter()
	srv := &http.Server{
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		Handler:      r,
		Addr:         ":6662",
	}
	log.Println("Starting server...")
	log.Fatal(srv.ListenAndServe())
}
