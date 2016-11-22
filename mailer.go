//  Iris Auth Manager
//
//  Created by Sliwa, Robert.
//  Copyright © 2016 Comcast Inc. All rights reserved.
//
package main

import (
	"bytes"
	"log"
	"net/http"
	"net/smtp"
	"os"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

func main() {
	serverPort := "4601"
	r := mux.NewRouter().StrictSlash(true)
	r.HandleFunc("/sendemail", handleSendEmail)
	if err := http.ListenAndServe(":"+serverPort, handlers.LoggingHandler(os.Stdout, r)); err != nil {
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
	// Connect to the remote SMTP server.
	c, err := smtp.Dial("mailrelay.comcast.com:25")
	if err != nil {
		log.Println(err)
		respondErr(w, r, http.StatusUnauthorized, err)
		return
	}
	defer c.Close()
	// Set the sender and recipient.
	c.Mail("irisadmin@iris.comcast.com")
	c.Rcpt("robert_sliwa@comcast.com")
	// Send the email body.
	wc, err := c.Data()
	if err != nil {
		log.Println(err)
		respondErr(w, r, http.StatusBadRequest, err)
		return
	}
	defer wc.Close()
	buf := bytes.NewBufferString("This is the email body.")
	if _, err = buf.WriteTo(wc); err != nil {
		log.Println(err)
		respondErr(w, r, http.StatusBadRequest, err)
		return
	}

	respond(w, r, http.StatusOK, map[string]interface{}{
		"message": "success",
	})
}
