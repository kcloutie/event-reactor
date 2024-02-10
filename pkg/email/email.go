package email

import (
	"bytes"
	"context"
	"fmt"
	"net/smtp"
	"time"

	"go.uber.org/zap"
)

var MimeHeaders = "MIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";\n\n"
var EmailTemplate string = `
<!DOCTYPE html>
<html>

<head>
  <style>
    body {
      box-sizing: border-box;
      min-width: 200px;
      max-width: 800px;
      margin: 0 auto;
      padding: 45px;
      font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Helvetica, Arial, sans-serif, "Apple Color Emoji", "Segoe UI Emoji";
    }

    h1,
    h2,
    h3,
    h4,
    h5,
    h6 {
      margin-top: 24px;
      margin-bottom: 16px;
      font-weight: 600;
      line-height: 1.25;
    }

    h1 {
      margin: .67em 0;
      font-weight: 600;
      padding-bottom: .3em;
      font-size: 2em;
      border-bottom: 1px solid hsla(210, 18%, 87%, 1);
    }

    h2 {
      font-weight: 600;
      padding-bottom: .3em;
      font-size: 1.5em;
      border-bottom: 1px solid hsla(210, 18%, 87%, 1);
    }

    table {
      border: 1px solid #1C6EA4;
      background-color: #EEEEEE;

      text-align: left;
      border-collapse: collapse;
    }

    table td,
    th {
      border: 1px solid #AAAAAA;
      padding: 3px 2px;
      font-size: 18px;
    }

    table tbody td {
      font-size: 16px;
      color: #333333;
    }

    table tr:nth-child(2n) {
      background-color: #f6f8fa;
    }
  </style>
</head>
<body>
  {{BODY}}
</body>
</html>
`

type EmailConfiguration struct {
	Log           *zap.Logger
	From          string
	Password      string
	To            []string
	Subject       string
	HtmlBody      string
	SMTPHost      string
	SMTPPort      int
	MaxRetries    int
	SleepInterval time.Duration
}

func New(from string, password string, to []string, subject string, htmlBody string) EmailConfiguration {
	return EmailConfiguration{
		From:          from,
		Password:      password,
		To:            to,
		Subject:       subject,
		HtmlBody:      htmlBody,
		SMTPHost:      "mrls.azell.com",
		SMTPPort:      587,
		MaxRetries:    30,
		SleepInterval: time.Duration(1) * time.Second,
	}
}

func (e *EmailConfiguration) SendEmail(ctx context.Context) error {
	log := e.Log.Sugar()
	auth := smtp.PlainAuth("", e.From, e.Password, e.SMTPHost)

	var body bytes.Buffer

	body.Write([]byte(fmt.Sprintf("Subject: %s \n%s\n\n%s", e.Subject, MimeHeaders, e.HtmlBody)))

	smtpHostFullName := fmt.Sprintf("%s:%v", e.SMTPHost, e.SMTPPort)

	var sendMailError error = nil
	for i := 1; i < e.MaxRetries+1; i++ {

		sendMailError = smtp.SendMail(smtpHostFullName, auth, e.From, e.To, body.Bytes())

		if sendMailError != nil {

			log.Debugf("Attempt %v of %v failed...sleeping and trying again", i, e.MaxRetries)
			time.Sleep(e.SleepInterval)
		} else {
			return nil
		}
	}

	return fmt.Errorf("failed to send email after %v attempts. Last error was %v", e.MaxRetries, sendMailError)
}
