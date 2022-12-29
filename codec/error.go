package codec

import "errors"

var (
	ErrInvalidSequence        = errors.New("invalid sequence number in response")
	ErrUnexpectedChecksum     = errors.New("unexpected checksum")
	ErrNotFoundCompressor     = errors.New("not found compressor")
	ErrCompressorTypeMismatch = errors.New("request and response Compressor type mismatch")
)
