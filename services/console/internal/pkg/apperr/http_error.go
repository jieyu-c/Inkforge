package apperr

import (
	"fmt"
)

// Body is serialized as JSON auth error payloads (PRD appendix B).
type Body struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// HTTP is an application error with stable machine code and HTTP status.
type HTTP struct {
	Status  int
	Code    string
	Message string
}

func (e *HTTP) Error() string {
	if e.Message != "" {
		return e.Message
	}
	return fmt.Sprintf("%s", e.Code)
}

func BadRequest(code, msg string) *HTTP {
	return &HTTP{Status: 400, Code: code, Message: msg}
}

func Unauthorized(code, msg string) *HTTP {
	return &HTTP{Status: 401, Code: code, Message: msg}
}

func Forbidden(code, msg string) *HTTP {
	return &HTTP{Status: 403, Code: code, Message: msg}
}

func NotFound(code, msg string) *HTTP {
	return &HTTP{Status: 404, Code: code, Message: msg}
}

func Conflict(code, msg string) *HTTP {
	return &HTTP{Status: 409, Code: code, Message: msg}
}

func TooMany(code, msg string) *HTTP {
	return &HTTP{Status: 429, Code: code, Message: msg}
}
