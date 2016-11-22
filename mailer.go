//  Iris Auth Manager
//
//  Created by Sliwa, Robert.
//  Copyright © 2016 Comcast Inc. All rights reserved.
//
package main

import (
	"bytes"
	"errors"
	"io"
	"log"
	"net/http"
	"net/smtp"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

// For graceful shutdown
func handle_signals() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	signal.Notify(c, syscall.SIGTERM)
	go func() {
		<-c
		log.Println("counter=emailRelayStop, Message=Shutting down email relay .... ")
		os.Exit(1)
	}()
}

// Read from configuration file and validate keys exist
func getConfiguration(config_file string) (err error) {
	config := EmailConfigurationInstance()
	return config.GetConfiguration(config_file)
}

func init() {
	if len(os.Args) < 2 {
		log.Fatal("Must pass in configuration file")
	}
	err := getConfiguration(os.Args[1])
	if err != nil {
		log.Println("Failed to get configuration file.")
		log.Fatal(err)
	}

	// start logging to file
	logFile, err := os.OpenFile(EmailConfigurationInstance().Log_file, os.O_CREATE|os.O_APPEND|os.O_RDWR, 0666)
	if err != nil {
		log.Fatal(err)
	}
	mw := io.MultiWriter(os.Stdout, logFile)
	log.SetOutput(mw)
}

func main() {
	handle_signals()
	r := mux.NewRouter().StrictSlash(true)
	r.HandleFunc("/sendemail", handleSendEmail)
	if err := http.ListenAndServe(EmailConfigurationInstance().Server_host+":"+EmailConfigurationInstance().Server_port, handlers.LoggingHandler(os.Stdout, r)); err != nil {
		log.Fatal(err)
	}
}

func handleSendEmail(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "POST":
		mailrelayhandler(w, r)
		return
	}

	// not found
	respondHTTPErr(w, r, http.StatusNotFound)
}

func mailrelayhandler(w http.ResponseWriter, r *http.Request) {
	// decode request body for user request key values
	var objectData map[string]interface{}
	err := decodeBody(r, &objectData)

	if err != nil {
		log.Println("Could not decode request body")
		respondErr(w, r, http.StatusBadRequest, errors.New("Could not decode request body"))
		return
	}

	toEmails, ok := objectData["to_emails"].([]interface{})
	if !ok {
		log.Println("Could not find to_emails in payload")
		respondErr(w, r, http.StatusBadRequest, errors.New("Could not find to_emails in payload"))
		return
	}

	var toEmailsStrings []string
	for _, email := range toEmails {
		emailString, ok := email.(string)
		if !ok {
			log.Println("to_emails array has wrong type")
			respondErr(w, r, http.StatusBadRequest, errors.New("to_emails array has wrong type"))
			return
		}
		toEmailsStrings = append(toEmailsStrings, emailString)
	}

	fromEmail, ok := objectData["from_email"].(string)
	if !ok {
		log.Println("Could not find from_email in payload")
		respondErr(w, r, http.StatusBadRequest, errors.New("Could not find from_email in payload"))
		return
	}

	subject, ok := objectData["subject"].(string)
	if !ok {
		log.Println("Could not find subject in payload")
		respondErr(w, r, http.StatusBadRequest, errors.New("Could not find subject in payload"))
		return
	}

	emailBody, ok := objectData["email_body"].(string)
	if !ok {
		log.Println("Could not find email_body in payload")
		respondErr(w, r, http.StatusBadRequest, errors.New("Could not find email_body in payload"))
		return
	}

	maxRetries, err := strconv.ParseInt(EmailConfigurationInstance().Max_retries, 10, 64)
	if err != nil {
		log.Println("Could not get max_retries from configuration file.")
		log.Println(err)
		respondErr(w, r, http.StatusUnauthorized, err)
		return
	}

	go func() {
		var currentAttempt int64 = 0
		isSuccess := false

		for isSuccess == false && currentAttempt < maxRetries {
			err := sendEmail(toEmailsStrings, fromEmail, subject, emailBody)
			if err == nil {
				log.Println("Email sent")
				isSuccess = true
			} else {
				currentAttempt += 1
			}
			if err != nil {
				log.Println("Error:", err)
			}
		}
	}()

	respond(w, r, http.StatusOK, map[string]interface{}{
		"message": "success",
	})
}

func sendEmail(toEmails []string, fromEmail, subject, body string) error {
	// Connect to the remote SMTP server.
	c, err := smtp.Dial(EmailConfigurationInstance().Mailrelay_fqdn)
	if err != nil {
		log.Println("Error dialing...")
		return err
	}
	defer c.Close()
	log.Println("Dialed and sending...")
	// Set the sender and recipient.
	c.Mail(fromEmail)
	for _, toEmail := range toEmails {
		c.Rcpt(toEmail)
	}

	// Send the email body.
	wc, err := c.Data()
	if err != nil {
		log.Println("Error executing DATA command...")
		return err
	}
	defer wc.Close()
	log.Println("Writing DATA...")
	buf := bytes.NewBufferString("To: " + toEmails[0] + "\nSubject: " + subject + "\nMIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";\n\n" + body)
	if _, err = buf.WriteTo(wc); err != nil {
		log.Println("Writing DATA...")
		return err
	}

	log.Println("DONE...")
	return nil
}
