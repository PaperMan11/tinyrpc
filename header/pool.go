package header

import "sync"

var (
	RequestPool  sync.Pool
	ResponsePool sync.Pool
)

func init() {
	RequestPool = sync.Pool{
		New: func() any {
			return &RequestHeader{}
		},
	}

	ResponsePool = sync.Pool{
		New: func() any {
			return &ResponseHeader{}
		},
	}
}
