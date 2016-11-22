//  Iris Auth Manager
//
//  Created by Sliwa, Robert.
//  Copyright © 2016 Comcast Inc. All rights reserved.
//
package main

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func decodeBody(r *http.Request, v interface{}) error {
	defer r.Body.Close()

	return json.NewDecoder(r.Body).Decode(v)
}

func encodeBody(w http.ResponseWriter, r *http.Request, v interface{}) error {
	return json.NewEncoder(w).Encode(v)
}

func respondErr(w http.ResponseWriter, r *http.Request, status int, args ...interface{}) {
	respond(w, r, status, map[string]interface{}{
		"error": map[string]interface{}{
			"message": fmt.Sprint(args...),
		},
	})
}

func respondErrWithCode(w http.ResponseWriter, r *http.Request, code string, status int, args ...interface{}) {
	respond(w, r, status, map[string]interface{}{
		"error": map[string]interface{}{
			"code":    code,
			"message": fmt.Sprint(args...),
		},
	})
}

func respond(w http.ResponseWriter, r *http.Request, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if data != nil {
		encodeBody(w, r, data)
	}
}

func respondHTTPErr(w http.ResponseWriter, r *http.Request, status int) {
	respondErr(w, r, status, http.StatusText(status))
}
