package main

import (
  "fmt"
  "net/http"
)

// Generic helper for logging an error message
func (app *application) logError(r *http.Request, err error) {
  app.logger.Println(err)
}

// Helper for sending json-formatted error messages to clients w/status code
func (app *application) errorResponse(w http.ResponseWriter, r *http.Request, status int, message interface{}) {
  env := envelope{"error": message}

  err := app.writeJSON(w, status, env, nil)
  if err != nil {
    app.logError(r, err)
    w.WriteHeader(500)
  }
}

// Helper when our app encounters an unexpected problem at runtime. Send 500
func (app *application) serverErrorResponse(w http.ResponseWriter, r *http.Request, err error) {
  app.logError(r, err)

  message := "The server encountered a proble and could not process your request"

  app.errorResponse(w, r, http.StatusInternalServerError, message)
}

// Helper when we encounter a 404
func (app *application) notFoundResponse(w http.ResponseWriter, r *http.Request) {
  message := "The requested resource could not be found"
  app.errorResponse(w, r, http.StatusNotFound, message)
}

// Helper when request is made with incorrect method
func (app *application) methodNotAllowedResponse(w http.ResponseWriter, r *http.Request) {
  message := fmt.Sprintf("the %s method is not supported for this resource", r.Method)
  app.errorResponse(w, r, http.StatusMethodNotAllowed, message)
}

// Helper for handling bad requests
func (app *application) badRequestResponse(w http.ResponseWriter, r *http.Request, err error) {
  app.errorResponse(w, r, http.StatusBadRequest, err.Error())
}

// Helper for failed JSON validation responses
func (app *application) failedValidationResponse(w http.ResponseWriter, r *http.Request, errors map[string]string) {
  app.errorResponse(w, r, http.StatusUnprocessableEntity, errors)
}

