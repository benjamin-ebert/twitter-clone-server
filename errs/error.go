package errs

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
)

// Application error codes.
//
// NOTE: These are meant to be generic and they map well to HTTP error codes.
// Different applications can have very different error code requirements so
// these should be expanded as needed (or introduce subcodes).
const (
	// Generic errors
	ECONFLICT       = "conflict"
	EINTERNAL       = "internal"
	EINVALID        = "invalid"
	ENOTFOUND       = "not_found"
	ENOTIMPLEMENTED = "not_implemented"
	EUNAUTHORIZED   = "unauthorized"
)

// lookup of application error codes to HTTP status codes.
var codes = map[string]int{
	ECONFLICT:       http.StatusConflict,
	EINVALID:        http.StatusBadRequest,
	ENOTFOUND:       http.StatusNotFound,
	ENOTIMPLEMENTED: http.StatusNotImplemented,
	EUNAUTHORIZED:   http.StatusUnauthorized,
	EINTERNAL:       http.StatusInternalServerError,
}

// ReturnError prints & optionally logs an error message.
func ReturnError(w http.ResponseWriter, r *http.Request, err error) {
	// Extract error code & message.
	code, message := ErrorCode(err), ErrorMessage(err)

	// Log & report internal errors.
	if code == EINTERNAL {
		//ReportError(r.Context(), err, r)
		LogError(r, err)
	}

	// Print user message to response based on reqeust accept header.
	w.Header().Set("Content-type", "application/json")
	w.WriteHeader(ErrorStatusCode(code))
	json.NewEncoder(w).Encode(&ErrorResponse{Error: message})
}

// Error represents an application-specific error. Application errors can be
// unwrapped by the caller to extract out the code & message.
//
// Any non-application error (such as a disk error) should be reported as an
// EINTERNAL error and the human user should only see "Internal error" as the
// message. These low-level internal error details should only be logged and
// reported to the operator of the application (not the end user).
type Error struct {
	// Machine-readable error code.
	Code string

	// Human-readable error message.
	Message string
}

// Error implements the error interface. Not used by the application otherwise.
func (e *Error) Error() string {
	return fmt.Sprintf("wtf error: code=%s message=%s", e.Code, e.Message)
}

// ErrorCode unwraps an application error and returns its code.
// Non-application errors always return EINTERNAL.
func ErrorCode(err error) string {
	var e *Error
	if err == nil {
		return ""
	} else if errors.As(err, &e) {
		return e.Code
	}
	return EINTERNAL
}

// ErrorMessage unwraps an application error and returns its message.
// Non-application errors always return "Internal error".
func ErrorMessage(err error) string {
	var e *Error
	if err == nil {
		return ""
	} else if errors.As(err, &e) {
		return e.Message
	}
	return "Internal error."
}

// Errorf is a helper function to return an Error with a given code and formatted message.
func Errorf(code string, format string, args ...interface{}) *Error {
	return &Error{
		Code:    code,
		Message: fmt.Sprintf(format, args...),
	}
}

// ErrorResponse represents a JSON structure for error output.
type ErrorResponse struct {
	Error string `json:"error"`
}

// parseResponseError parses an JSON-formatted error response.
func parseResponseError(resp *http.Response) error {
	defer resp.Body.Close()

	// Read the response body so we can reuse it for the error message if it
	// fails to decode as JSON.
	buf, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	// Parse JSON formatted error response.
	// If not JSON, use the response body as the error message.
	var errorResponse ErrorResponse
	if err := json.Unmarshal(buf, &errorResponse); err != nil {
		message := strings.TrimSpace(string(buf))
		if message == "" {
			message = "Empty response from server."
		}
		return Errorf(FromErrorStatusCode(resp.StatusCode), message)
	}
	return Errorf(FromErrorStatusCode(resp.StatusCode), errorResponse.Error)
}

// LogError logs an error with the HTTP route information.
func LogError(r *http.Request, err error) {
	log.Printf("[http] error: %s %s: %s", r.Method, r.URL.Path, err)
}

// ErrorStatusCode returns the associated HTTP status code for a WTF error code.
func ErrorStatusCode(code string) int {
	if v, ok := codes[code]; ok {
		return v
	}
	return http.StatusInternalServerError
}

// FromErrorStatusCode returns the associated WTF code for an HTTP status code.
func FromErrorStatusCode(code int) string {
	for k, v := range codes {
		if v == code {
			return k
		}
	}
	return EINTERNAL
}



const (
	// RememberRequired is returned when a create or update
	// is attempted without a user remember token hash
	RememberRequired privateError = "models: remember token is required"
	// RememberTooShort is returned when a remember token is
	// not at least 32 bytes
	RememberTooShort privateError = "models: remember token must be at least 32 bytes"
)

type privateError string

func (e privateError) Error() string {
	return string(e)
}
