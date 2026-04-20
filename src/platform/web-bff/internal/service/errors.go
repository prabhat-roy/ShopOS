package service

import "errors"

var (
	ErrNotFound       = errors.New("not found")
	ErrUnauthorized   = errors.New("unauthorized")
	ErrInvalidInput   = errors.New("invalid input")
	ErrNotImplemented = errors.New("gRPC client not yet implemented — pending proto compilation")
	ErrUpstream       = errors.New("upstream service error")
)
