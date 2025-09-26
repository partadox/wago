package error

import "net/http"

// GenericError represent as the contract of generic error
type GenericError interface {
	Error() string
	ErrCode() string
	StatusCode() int
}

type InternalServerError string

// Error for complying the error interface
func (e InternalServerError) Error() string {
	return string(e)
}

// ErrCode will return the error code based on the error data type
func (e InternalServerError) ErrCode() string {
	return "INTERNAL_SERVER_ERROR"
}

// StatusCode will return the HTTP status code based on the error data type
func (e InternalServerError) StatusCode() int {
	return http.StatusInternalServerError
}

type ContextError string

// Error for complying the error interface
func (e ContextError) Error() string {
	return string(e)
}

// ErrCode will return the error code based on the error data type
func (e ContextError) ErrCode() string {
	return "CONTEXT_ERROR"
}

// StatusCode will return the HTTP status code based on the error data type
func (e ContextError) StatusCode() int {
	return http.StatusRequestTimeout
}

type BadRequestError string

// Error for complying the error interface
func (e BadRequestError) Error() string {
	return string(e)
}

// ErrCode will return the error code based on the error data type
func (e BadRequestError) ErrCode() string {
	return "BAD_REQUEST"
}

// StatusCode will return the HTTP status code based on the error data type
func (e BadRequestError) StatusCode() int {
	return http.StatusBadRequest
}

type NotFoundError string

// Error for complying the error interface
func (e NotFoundError) Error() string {
	return string(e)
}

// ErrCode will return the error code based on the error data type
func (e NotFoundError) ErrCode() string {
	return "NOT_FOUND"
}

// StatusCode will return the HTTP status code based on the error data type
func (e NotFoundError) StatusCode() int {
	return http.StatusNotFound
}

type RequestTimeoutError string

// Error for complying the error interface
func (e RequestTimeoutError) Error() string {
	return string(e)
}

// ErrCode will return the error code based on the error data type
func (e RequestTimeoutError) ErrCode() string {
	return "REQUEST_TIMEOUT"
}

// StatusCode will return the HTTP status code based on the error data type
func (e RequestTimeoutError) StatusCode() int {
	return http.StatusRequestTimeout
}
