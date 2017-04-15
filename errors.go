package main

import (
	"fmt"
	"log"
	"net/http"
)

func writeError(w http.ResponseWriter, msg string, err error) {
	var theError string
	if err != nil {
		theError = fmt.Sprintf("%s: %v", msg, err)
	} else {
		theError = msg
	}
	log.Print(theError)
	http.Error(w, theError, 400)
}

var ErrBraceletNumberDoesNotExist = "bracelet number does not exist"
var ErrNoBraceletNumberEntered = "no bracelet number was entered"
var ErrBraceletNumberInvalid = "invalid bracelet number"
var ErrNoAdmissionTypeSelected = "admission type not selected. please select an admission type"

var ErrSessionAlreadyExists = "session already exists"
var ErrSessionDoesNotExist = "session does not exist"

var ErrParsingTemplate = "could not parse template"
var ErrExecutingTemplate = "could not execute template"

var ErrWithSQLquery = "could not perform query"

var ErrItemsOnlyWithLockerZero = "the items only box can only be used with locker number 0 (zero)"
