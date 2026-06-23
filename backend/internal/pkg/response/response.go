package response

import (
	"encoding/json"
	"net/http"
)

const (
	CodeOK              = 0
	CodeBadRequest      = 40001
	CodeSensitiveBlock  = 40002
	CodeForbidden       = 40301
	CodeNotFound        = 40401
	CodeTooManyRequests = 42901
	CodeInternalError   = 50000
)

type Body struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

func JSON(w http.ResponseWriter, httpStatus int, code int, message string, data any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(httpStatus)
	_ = json.NewEncoder(w).Encode(Body{
		Code:    code,
		Message: message,
		Data:    data,
	})
}

func OK(w http.ResponseWriter, data any) {
	JSON(w, http.StatusOK, CodeOK, "ok", data)
}

func Error(w http.ResponseWriter, httpStatus, code int, message string) {
	JSON(w, httpStatus, code, message, nil)
}
