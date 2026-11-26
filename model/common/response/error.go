package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/mo7amed4522/tayara_api/constants"
)

// AppError is a platform error with a machine-readable code and an HTTP status. Services may
// return it; the API layer maps it through FailWithError so every failure carries the same
// {code, data, msg} envelope with a truthful HTTP status.
type AppError struct {
	Code  int    // machine-readable code (see constants/errors.go)
	Msg   string `json:"msg"`
	HTTP  int    // HTTP status
	cause error  // optional wrapped cause (never serialized)
}

func (e *AppError) Error() string { return e.Msg }

// WithCode wraps a plain message into a client-visible AppError (default internal server error).
func WithCode(code int, msg string) *AppError {
	return &AppError{Code: code, Msg: msg, HTTP: httpStatusFor(code)}
}

// Wrap attaches the code class + message for a Go error, preserving its text for logging.
func Wrap(code int, err error) *AppError {
	if err == nil {
		return nil
	}
	return &AppError{Code: code, Msg: err.Error(), HTTP: httpStatusFor(code), cause: err}
}

// ValidationErr is a shorthand for request-validation failures (HTTP 422).
func ValidationErr(msg string) *AppError { return WithCode(constants.CodeValidation, msg) }

// NotfoundErr is a shorthand for missing resources (HTTP 404).
func NotFoundErr(msg string) *AppError { return WithCode(constants.CodeNotFound, msg) }

// ConflictErr is a shorthand for state/duplicate conflicts (HTTP 409).
func ConflictErr(msg string) *AppError { return WithCode(constants.CodeConflict, msg) }

// ForbiddenErr is a shorthand for authenticated-but-not-allowed failures (HTTP 403).
func ForbiddenErr(msg string) *AppError { return WithCode(constants.CodeForbidden, msg) }

// InternalErr is a shorthand for unexpected failures (HTTP 500).
func InternalErr(msg string) *AppError { return WithCode(constants.CodeInternal, msg) }

func httpStatusFor(code int) int {
	switch code {
	case constants.CodeUnauthorized:
		return http.StatusUnauthorized
	case constants.CodeForbidden:
		return http.StatusForbidden
	case constants.CodeNotFound:
		return http.StatusNotFound
	case constants.CodeConflict:
		return http.StatusConflict
	case constants.CodeValidation:
		return http.StatusUnprocessableEntity
	case constants.CodeProvider:
		return http.StatusBadGateway
	default:
		return http.StatusInternalServerError
	}
}

// FailWithError sends an AppError with its registered HTTP status and the standard envelope.
func FailWithError(err *AppError, c *gin.Context) {
	if err == nil {
		Fail(c)
		return
	}
	c.JSON(err.HTTP, Response{
		Code: err.Code,
		Data: map[string]interface{}{},
		Msg:  err.Msg,
	})
}

// FailWithCode sends a code + message at the code's derived HTTP status (uniform envelope).
func FailWithCode(code int, msg string, c *gin.Context) {
	c.JSON(httpStatusFor(code), Response{
		Code: code,
		Data: map[string]interface{}{},
		Msg:  msg,
	})
}