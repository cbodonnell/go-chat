package main

import "net/http"

func badRequest(w http.ResponseWriter, err error) {
	msg := err.Error()
	http.Error(w, msg, http.StatusBadRequest)
}

func unauthorizedRequest(w http.ResponseWriter, err error) {
	msg := err.Error()
	http.Error(w, msg, http.StatusUnauthorized)
}

func internalServerError(w http.ResponseWriter, err error) {
	msg := err.Error()
	http.Error(w, msg, http.StatusInternalServerError)
}
