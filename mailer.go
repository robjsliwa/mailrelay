//  Iris Auth Manager
//
//  Created by Sliwa, Robert.
//  Copyright © 2016 Comcast Inc. All rights reserved.
//
package main

import (
	"bytes"
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
		log.Println("counter=authMgrStop, Message=Shutting down auth manager .... ")
		os.Exit(1)
	}()
}

// Read from configuration file and validate keys exist
func getConfiguration(config_file string) (err error) {
	config := EmailConfigurationInstance()
	return config.GetConfiguration(config_file)
}

func init() {
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
	/*

		  log.Println(err)
		  respondErr(w, r, http.StatusUnauthorized, err)
		  return

		  respond(w, r, http.StatusOK, map[string]interface{}{
				"message": "success",
			})

	*/
	maxRetries, err := strconv.ParseInt(EmailConfigurationInstance().Max_retries, 10, 64)
	if err != nil {
		log.Println("Could not get max_retries from configuration file.")
		log.Println(err)
		respondErr(w, r, http.StatusUnauthorized, err)
		return
	}

	var currentAttempt int64 = 0
	isSuccess := false

	for isSuccess == false && currentAttempt < maxRetries {
		err := sendEmail([]string{"robert_sliwa@comcast.com"}, "donotreply@iris.comcast.com", "Test email", "Test test")
		if err == nil {
			log.Println("Email sent")
			isSuccess = true
		} else {
			currentAttempt += 1
		}
        log.Println("Error:", err)
	}
}

func sendEmail(toEmails []string, fromEmail, subject, body string) error {
	// Connect to the remote SMTP server.
	c, err := smtp.Dial("mailrelay.comcast.com:25")
	if err != nil {
        log.Println("Error dialing...")
		return err
	}
	defer c.Close()
    log.Println("Dialed and sending...")
	// Set the sender and recipient.
	c.Mail("irisadmin@iris.comcast.com")
	c.Rcpt("robert_sliwa@comcast.com")
	// Send the email body.
	wc, err := c.Data()
	if err != nil {
        log.Println("Error executing DATA command...")
		return err
	}
	defer wc.Close()
    log.Println("Writing DATA...")
	buf := bytes.NewBufferString("To: robert_sliwa@comcast.com\nSubject: Test with subject\nMIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";\n\n<html><body><h1>Test message in html</h1></body></html>")
	if _, err = buf.WriteTo(wc); err != nil {
        log.Println("Writing DATA...")
		return err
	}

    log.Println("DONE...")
	return nil
}
