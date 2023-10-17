package controllers

import (
	"errors"
)

var (
	errNotExist  = errors.New("does not exist")
	errInvalidID = errors.New("error parsing id parameter")
	errIDLatest  = errors.New("latest id requested")
	errIDNew     = errors.New("new")
	// errInsertFailed = errors.New("Error in the user insertion")
	// errUpdateFailed = errors.New("Error in the user updation")
	// errDeleteFailed = errors.New("Error in the user deletion")
)

func getErrorStrings(errors []error) (errorStrings []string) {
	for _, err := range errors {
		errorStrings = append(errorStrings, err.Error())
	}
	return
}
