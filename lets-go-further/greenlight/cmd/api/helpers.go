package main

import (
  "io"
  "fmt"
  "errors"
  "strings"
  "net/http"
  "strconv"
  "encoding/json"
  "github.com/julienschmidt/httprouter"
)

func (app *application) readIDParam(r *http.Request) (int64, error) {
  params := httprouter.ParamsFromContext(r.Context())

  id, err := strconv.ParseInt(params.ByName("id"), 10, 64)
  if err != nil {
    return 0, errors.New("invalid id parameter")
  }

  return id, nil
}

type envelope map[string]interface{}

func (app *application) writeJSON(w http.ResponseWriter, status int, data envelope, headers http.Header) error {
  js, err := json.MarshalIndent(data, "", "\t")
  if err != nil {
    return err
  }

  js = append(js, '\n')

  for k, v := range headers {
    w.Header()[k] = v
  }

  w.Header().Set("Content-Type", "application/json")
  w.WriteHeader(status)
  w.Write(js)
  return nil
}


// Helper decode JSON from request body and replace with our own error messages
func (app *application) readJSON(w http.ResponseWriter, r *http.Request, dst interface{}) error {
  // restrict the size of the request body
  maxBytes := 1_048_576
  r.Body = http.MaxBytesReader(w, r.Body, int64(maxBytes))

  dec := json.NewDecoder(r.Body)
  dec.DisallowUnknownFields()
  err := dec.Decode(dst)

  if err != nil {
    // If there is an error during decoding, start the triage...
    var syntaxError *json.SyntaxError
    var unmarshalTypeError *json.UnmarshalTypeError
    var invalidUnmarshalError *json.InvalidUnmarshalError

    switch {
    case errors.As(err, &syntaxError):
      return fmt.Errorf("body contains badly-formed JSON at character %d", syntaxError.Offset)

    case errors.Is(err, io.ErrUnexpectedEOF):
      return errors.New("body contains badly-formed JSON")

    case errors.As(err, &unmarshalTypeError):
      if unmarshalTypeError.Field != "" {
        return fmt.Errorf("body contains incorrect JSON type for field %q", unmarshalTypeError.Field)
      }
      return fmt.Errorf("body conaints incorrect JSON type")

    case errors.Is(err, io.EOF):
      return errors.New("body must not be empty")

    case strings.HasPrefix(err.Error(), "json: unknown field "):
      fieldName := strings.TrimPrefix(err.Error(), "json: unknown field ")
      return fmt.Errorf("body contains unknown key %s", fieldName)

    case err.Error() == "http: request body too large":
      return fmt.Errorf("body must not ne larger than max size")

    case errors.As(err, &invalidUnmarshalError):
      panic(err)

    default:
      return err
    }
  }

  err = dec.Decode(&struct{}{})
  if err != io.EOF {
    return errors.New("body must only contain single JSON value")
  }

  return nil
}




















