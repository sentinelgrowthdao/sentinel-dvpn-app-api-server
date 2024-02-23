package middleware

import "errors"

type APIError error

var (
	APIErrorUnknown        APIError = errors.New("unknownError")
	APIErrorInvalidRequest APIError = errors.New("invalidRequest")
	APIErrorNotFound       APIError = errors.New("notFound")
)
