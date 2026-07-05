package handshake

import "errors"

var (
	ErrInvalidVersion = errors.New("RTMP version is invalid, expected 3")
)