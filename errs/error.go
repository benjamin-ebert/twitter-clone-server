package errs

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
)

// Generic errors that map well to common HTTP status codes.
// They are returned on most error occasions.
// More specific errors are defined at the bottom.
const (
	ECONFLICT       = "conflict"
	EINTERNAL       = "internal"
	EINVALID        = "invalid"
	ENOTFOUND       = "not_found"
	EUNAUTHORIZED   = "unauthorized"
)

// Mapping of error codes to HTTP status codes.
var codes = map[string]int{
	ECONFLICT:       http.StatusConflict,
	EINVALID:        http.StatusBadRequest,
	ENOTFOUND:       http.StatusNotFound,
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

	// Print user message to response.
	w.WriteHeader(ErrorStatusCode(code))
	json.NewEncoder(w).Encode(&ErrorResponse{Error: message})
}

// Error represents an app-specific error.
// They can be unwrapped by the caller to extract the code and the message.
// App-agnostic errors (like error bcrypting a password) are reported as
// error EINTERNAL. The user only sees the message "Internal error".
// Details of internal errors are only logged and reported to the operator.
type Error struct {
	// Machine-readable error code.
	Code string
	// Human-readable error message.
	Message string
}

// Error implements the error interface. Not used by the app otherwise.
func (e *Error) Error() string {
	return fmt.Sprintf("wtf error: code=%s message=%s", e.Code, e.Message)
}

// ErrorCode unwraps an application error and returns its code.
// App-agnostic errors always return EINTERNAL.
func ErrorCode(err error) string {
	var e *Error
	if err == nil {
		return ""
	} else if errors.As(err, &e) {
		return e.Code
	}
	return EINTERNAL
}

// ErrorMessage unwraps an app error and returns its message.
// App-agnostic errors always return "Internal error".
func ErrorMessage(err error) string {
	var e *Error
	if err == nil {
		return ""
	} else if errors.As(err, &e) {
		return e.Message
	}
	return "Internal error."
}

// Errorf is a helper to return an Error with a given code and formatted message.
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

// LogError logs an error with the HTTP route information.
func LogError(r *http.Request, err error) {
	log.Printf("[http] error: %s %s: %s", r.Method, r.URL.Path, err)
}

// ErrorStatusCode returns the associated HTTP status code for an app error code.
// See const "codes" above for the mapping.
func ErrorStatusCode(code string) int {
	if v, ok := codes[code]; ok {
		return v
	}
	return http.StatusInternalServerError
}

// Specific private errors that should not be reported to the user but should be reported
// to the operator. If they are returned, the user sees code 500 and "Internal error", and
// the actual message gets printed to the log.
const (
	// RememberHashEmpty is returned when a user-create or -update is attempted without a remember token hash.
	RememberHashEmpty privateError = "auth: the user's remember hash is an empty string."
	// RememberTooShort is returned when a remember token is shorter than 32 bytes.
	RememberTooShort privateError = "auth: the user's remember token must be at least 32 bytes."
)

type privateError string

func (e privateError) Error() string {
	return string(e)
}
