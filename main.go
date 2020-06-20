package main

import (
	"bytes"
	"database/sql"
	"encoding/csv"
	"fmt"
	"github.com/badoux/checkmail"
	"github.com/fridgecow/smartalarm-server/sleepdata"
	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
	"github.com/wcharczuk/go-chart"
	gomail "gopkg.in/mail.v2"
	"html/template"
	"io"
	"log"
	"math"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"time"
)

var db *sql.DB

var templates *template.Template

type EmailInfo struct {
	Id           int
	RegisterDate time.Time
	Email        string
	Status       string
	Token        string
	Uses         int
	LastUse      time.Time
}

type ConfirmData struct {
	Email string
	Token string
}

type SummaryEmail struct {
	Statistics sleepdata.SleepStatistics
	Email      string
	Token      string
}

var Err = fmt.Errorf

func generateToken() string {
	const letters = "0123456789abcdef"
	b := make([]byte, 64)

	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}

	return string(b)
}

func validateEmail(email string) bool {
	log.Println("Validating email", email)

	err := checkmail.ValidateFormat(email)
	if err != nil {
		return false
	}

	err = checkmail.ValidateHost(email)
	if err != nil {
		return false
	}

	err = checkmail.ValidateHost(email)
	if _, ok := err.(checkmail.SmtpError); ok && err != nil {
		return false
	}

	return true
}

type AttachmentContainer struct {
	name   string
	reader io.Reader
}

func handlerNoLog(f func(*http.Request) (string, error)) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")

		out, err := f(r)
		if err != nil {
			w.Write([]byte("Error: " + err.Error()))
			log.Println("Error:", err.Error())
			return
		}

		w.Write([]byte(out))
	}
}

func handler(f func(*http.Request) (string, error)) func(http.ResponseWriter, *http.Request) {
	handlerF := handlerNoLog(f)
	return func(w http.ResponseWriter, r *http.Request) {
		log.Println("Handling", r.Method, r.URL)
		handlerF(w, r)
	}
}

func sendEmail(to string, subject string, f func(io.Writer) error, files ...AttachmentContainer) error {
	msg := gomail.NewMessage()

	msg.SetAddressHeader("From", os.Getenv("SMARTALARM_EMAILADDR"), "Smart Alarm")
	msg.SetHeader("To", to)
	msg.SetHeader("ReplyTo", os.Getenv("SMARTALARM_EMAILREPLY"))

	msg.SetHeader("Subject", subject)
	msg.SetBody("text/plain", "Thank you for using Smart Alarm. Please switch to an HTML-capable client to view this email.")
	msg.AddAlternativeWriter("text/html", f)

	for _, file := range files {
		msg.EmbedReader(file.name, file.reader)
	}

	// Send on seperate channel
	emailChan <- msg

	return nil
}

func getEmailInfo(email string) (emailInfo EmailInfo, e error) {
	// DB query
	err := db.QueryRow(
		"SELECT id, register_date, email, type, token, uses, last_use FROM smartalarm_emails WHERE email = ?",
		email,
	).Scan(&emailInfo.Id, &emailInfo.RegisterDate, &emailInfo.Email, &emailInfo.Status, &emailInfo.Token, &emailInfo.Uses, &emailInfo.LastUse)

	if err != nil {
		return emailInfo, err
	}

	return emailInfo, nil
}

func addEmail(r *http.Request) (string, error) {
	params := mux.Vars(r)

	email := params["email"]

	info, err := getEmailInfo(email)
	if err == nil && info.Status == "CONFIRMED" {
		return "", Err("email already confirmed")
	}

	token := generateToken()
	if err != nil {
		// No results - add email to DB
		if !validateEmail(email) {
			return "", Err("email not valid")
		}

		log.Println("Registering", email)

		_, err := db.Exec(
			"INSERT INTO smartalarm_emails (email, token) VALUES (?, ?)",
			email,
			token,
		)
		if err != nil {
			return "", err
		}
	} else if info.Status == "UNSUBSCRIBED" || info.Status == "REGISTERED" {
		// Update token
		_, err := db.Exec(
			"UPDATE `smartalarm_emails` SET `token` = ? WHERE `email` = ?",
			token,
			email,
		)

		if err != nil {
			return "", err
		}
	}

	err = sendEmail(email, "Smart Alarm Email Confirmation", func(w io.Writer) error {
		return templates.Lookup("emailconfirm.tmpl").ExecuteTemplate(w, "confirmation", ConfirmData{Email: email, Token: token})
	})

	if err != nil {
		return "", err
	}

	return "Confirmation sent - check your email", nil
}

func confirmEmail(r *http.Request) (string, error) {
	params := mux.Vars(r)
	email := params["email"]
	token := params["token"]

	info, err := getEmailInfo(email)
	if err != nil {
		return "", Err("email not registered")
	}

	if info.Status == "CONFIRMED" {
		return "", Err("email already confirmed")
	}

	if info.Token != token {
		return "", Err("token doesn't match")
	}

	_, err = db.Exec(
		"UPDATE `smartalarm_emails` SET `type`='CONFIRMED', `token` = '' WHERE `email` = ?",
		email,
	)
	if err != nil {
		return "", err
	}

	log.Println("Confirmed", email)
	return "Email confirmed - you can now export your data.", nil
}

func unsubEmail(r *http.Request) (string, error) {
	params := mux.Vars(r)
	email := params["email"]
	token := params["token"]

	info, err := getEmailInfo(email)
	if err != nil {
		return "", Err("email not registered")
	}

	if info.Status != "CONFIRMED" {
		return "", Err("email not subscribed")
	}

	if info.Token != token {
		return "", Err("token doesn't match. Use most recent email.")
	}

	_, err = db.Exec(
		"UPDATE `smartalarm_emails` SET `type`='UNSUBSCRIBED', `token` = '' WHERE `email` = ?",
		email,
	)
	if err != nil {
		return "", err
	}

	log.Println("Unsubscribed", email)
	return "Email unsubscribed. To re-subscribe, send a new confirmation from your watch.", nil
}

func sendCSV(r *http.Request) (string, error) {
	params := mux.Vars(r)
	email := params["email"]
	tz := r.FormValue("tz")

	// Validate email and regenerate token before doing any other work
	info, err := getEmailInfo(email)
	if err != nil {
		return "", Err("email not registered")
	}

	if info.Status != "CONFIRMED" {
		return "", Err("email not subscribed")
	}

	info.Token = generateToken()
	_, err = db.Exec(
		"UPDATE `smartalarm_emails` SET `token` = ? WHERE `email` = ?",
		info.Token,
		email,
	)
	if err != nil {
		return "", Err(err.Error() + " db")
	}

	location, err := time.LoadLocation(tz)
	if err != nil {
		return "", Err(err.Error() + " tz")
	}

	csvData := csv.NewReader(strings.NewReader(r.FormValue("csv")))
	csvData.FieldsPerRecord = -1

	header, err := csvData.Read()
	if err != nil {
		return "", Err("can not read CSV: " + err.Error())
	}

	csvBody, err := csvData.ReadAll()
	if err != nil {
		return "", Err("can not read CSV body: " + err.Error())
	}

	var sleepSummary sleepdata.SleepSummary
	fullExport := header[0] != "Region Type"
	if fullExport {
		// Full Export - load sleep data then summarise it
		sleepData, err := sleepdata.MakeSleepData(csvBody, *location)
		if err != nil {
			return "", Err(err.Error() + " sd")
		}

		sleepSummary, err = sleepdata.SummariseData(sleepData)
		if err != nil {
			return "", Err(err.Error() + " ss")
		}

		log.Println("Full summary export")
	} else {
		// Summary export - parse sleep regions
		sleepSummary, err = sleepdata.ParseRegions(csvBody, location)
		if err != nil {
			return "", Err(err.Error() + " ss reg")
		}
	}

	graph := chart.Chart{
		Title:      sleepSummary.Title,
		TitleStyle: chart.StyleShow(),

		Width: 1200,
		Background: chart.Style{
			Padding: chart.Box{
				Left: 50,
			},
		},

		XAxis: chart.XAxis{
			Name:           "Time",
			NameStyle:      chart.StyleShow(),
			Style:          chart.StyleShow(),
			ValueFormatter: chart.TimeValueFormatterWithFormat("15:04"),
		},
		YAxis: chart.YAxis{
			Name:      "Motion (m/s²)",
			NameStyle: chart.StyleShow(),
			Style:     chart.StyleShow(),
		},
		YAxisSecondary: chart.YAxis{
			Name:      "Heart Rate (bpm)",
			NameStyle: chart.StyleShow(),
			Style:     chart.StyleShow(),
		},
	}

	if !fullExport {
		graph.YAxis.Range = &chart.ContinuousRange{
			Min: 0,
			Max: 1000,
		}

		graph.YAxis.Style.Show = false
		graph.YAxisSecondary.Style.Show = false

		graph.Series = sleepSummary.GetChartBands()
	} else {
		graph.Series = append(
			sleepSummary.GetChartBands(),
			sleepSummary.Data.GetMotionSeries(),
			sleepSummary.Data.GetHeartRateSeries(),
		)
	}

	if !sleepSummary.Statistics.HeartRateEnabled {
		graph.YAxisSecondary.Style.Show = false
	}

	graph.Elements = []chart.Renderable{
		chart.Legend(&graph),
	}

	buffer := bytes.NewBuffer([]byte{})
	err = graph.Render(chart.PNG, buffer)
	if err != nil {
		return "", Err("could not render graph: " + err.Error())
	}

	emailData := SummaryEmail{
		Token:      info.Token,
		Email:      info.Email,
		Statistics: sleepSummary.Statistics,
	}

	sendEmail(
		email,
		sleepSummary.Title,
		func(w io.Writer) error {
			return templates.Lookup("emailhtml.tmpl").ExecuteTemplate(w, "export", emailData)
		},
		AttachmentContainer{
			"image.png",
			buffer,
		},
		AttachmentContainer{
			"data.csv",
			bytes.NewBufferString(r.FormValue("csv")),
		},
	)

	return "Export success", nil
}

func main() {
	var err error

	// Database Connection
	db, err = sql.Open(
		"mysql",
		os.Getenv("SMARTALARM_DBUSER")+":"+os.Getenv("SMARTALARM_DBPASS")+"@"+os.Getenv("SMARTALARM_DBHOST")+"/"+os.Getenv("SMARTALARM_DBNAME")+"?parseTime=true",
	)
	if err != nil {
		log.Fatalf("Error on initializing database connection: %s", err.Error())
	}

	err = db.Ping()
	if err != nil {
		log.Fatalf("Error on opening database connection: %s", err.Error())
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
	r := mux.NewRouter()

	r.HandleFunc("/v1/add/{email}", handler(addEmail))
	r.HandleFunc("/v1/confirm/{email}/{token:[0-9a-f]{64,}}", handler(confirmEmail))
	r.HandleFunc("/v1/unsub/{email}/{token:[0-9a-f]{64,}}", handler(unsubEmail))
	r.HandleFunc("/v1/csv/{email}", handler(sendCSV))

	r.HandleFunc("/ping", handlerNoLog(func(r *http.Request) (string, error) { return "☑", nil }))

	srv := &http.Server{
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		Handler:      r,
		Addr:         ":6662",
	}
	log.Println("Starting server...")
	log.Fatal(srv.ListenAndServe())
}
