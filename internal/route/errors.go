package route

import (
	"errors"
	"net/http"
)

// HTTPError is an error that carries the HTTP status and payload it should
// produce. Handlers can return one directly instead of writing a response
// themselves; the engine renders it automatically.
//
//	if user == nil {
//		return route.NotFoundError("that account does not exist")
//	}
type HTTPError struct {
	Status  int
	Message string
	Err     error
}

func (e *HTTPError) Error() string {
	if e.Err != nil {
		return e.Message + ": " + e.Err.Error()
	}
	return e.Message
}

func (e *HTTPError) Unwrap() error {
	return e.Err
}

// NewHTTPError builds an HTTPError for status with message.
func NewHTTPError(status int, message string) *HTTPError {
	return &HTTPError{Status: status, Message: message}
}

// Wrap attaches an underlying error for logging without exposing it to clients.
func (e *HTTPError) Wrap(err error) *HTTPError {
	e.Err = err
	return e
}

// BadRequestError builds a 400 HTTPError.
func BadRequestError(message string) *HTTPError { return NewHTTPError(http.StatusBadRequest, message) }

// UnauthorizedError builds a 401 HTTPError.
func UnauthorizedError(message string) *HTTPError {
	return NewHTTPError(http.StatusUnauthorized, message)
}

// ForbiddenError builds a 403 HTTPError.
func ForbiddenError(message string) *HTTPError { return NewHTTPError(http.StatusForbidden, message) }

// NotFoundError builds a 404 HTTPError.
func NotFoundError(message string) *HTTPError { return NewHTTPError(http.StatusNotFound, message) }

// UnprocessableEntityError builds a 422 HTTPError.
func UnprocessableEntityError(message string) *HTTPError {
	return NewHTTPError(http.StatusUnprocessableEntity, message)
}

// InternalServerError builds a 500 HTTPError wrapping the causing error.
// The message is shown to the client; err is only logged.
func InternalServerError(message string, err error) *HTTPError {
	return NewHTTPError(http.StatusInternalServerError, message).Wrap(err)
}

// StatusForError reports the HTTP status err would produce if returned from
// a handler: an HTTPError's own status, 200 for a nil error, or 500 for
// anything else. Middleware that logs before the response is written (the
// engine renders HTTPError responses after the middleware chain returns) can
// use this to report the real outcome instead of assuming 500.
func StatusForError(err error) int {
	if err == nil {
		return http.StatusOK
	}
	if httpErr, ok := errors.AsType[*HTTPError](err); ok {
		return httpErr.Status
	}
	return http.StatusInternalServerError
}

// respondError writes an HTTPError as JSON, falling back to a generic 500
// for any other error so internal details are never leaked to clients.
func respondError(c *Context, err error) {
	if httpErr, ok := errors.AsType[*HTTPError](err); ok {
		_ = c.JSON(httpErr.Status, Data{"error": httpErr.Message})
		return
	}
	http.Error(c.Response, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
}
