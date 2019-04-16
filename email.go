package main

import (
  "time"
	"gopkg.in/mail.v2"
  "os"
  "log"
)

var emailChan chan *mail.Message

func init(){
  emailChan = make(chan *mail.Message)

  go func() {
      d := mail.NewDialer("smtp.gmail.com", 587, "smartalarm@fridgecow.com", os.Getenv("SMARTALARM_EMAILPASS"))
      d.StartTLSPolicy = mail.MandatoryStartTLS

      var s mail.SendCloser
      var err error
      open := false
      for {
          select {
          case m, ok := <-emailChan:
              log.Println("Sending email", m)
              if !ok {
                  return
              }
              if !open {
                  if s, err = d.Dial(); err != nil {
                      panic(err)
                  }
                  open = true
              }
              if err := mail.Send(s, m); err != nil {
                  log.Println(err)
                  return
              }

              // Increment "use" counter in DB
              go func(emails []string){
                if len(emails) != 1 {
                  log.Println("Wrong number of emails, not updating DB")
                  return
                }

                _, err := db.Exec("UPDATE smartalarm_emails SET uses=uses+1 WHERE email=?", emails[0])
                if err != nil {
                  log.Println(err)
                }
              }(m.GetHeader("To"))
          // Close the connection to the SMTP server if no email was sent in
          // the last 5 seconds.
          case <-time.After(5 * time.Second):
              if open {
                  log.Println("Closing SMTP connection")
                  if err := s.Close(); err != nil {
                      log.Println(err)
                  }
                  open = false
              }
          }
      }
  }()
}
