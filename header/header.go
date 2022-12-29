package header

import (
	"encoding/binary"
	"errors"
	"sync"
	"tinyrpc/compressor"
)

const (
	// MaxHeaderSize = 2 + 10 + 10 + 10 + 4 (10 refer to binary.MaxVarintLen64)
	MaxHeaderSize = 36
	Uint32Size    = 4 // byte
	Uint16Size    = 2
)

var ErrUnmarshal = errors.New("unmarshal error")

// RequestHeader request header structure looks like:
// 	+--------------+----------------+----------+------------+----------+
// 	| CompressType |      Method    |    ID    | RequestLen | Checksum |
// 	+--------------+----------------+----------+------------+----------+
// 	|    uint16    | uvarint+string |  uvarint |   uvarint  |  uint32  |
// 	+--------------+----------------+----------+------------+----------+
type RequestHeader struct {
	sync.RWMutex
	CompressType compressor.CompressType // 表示RPC的协议内容的压缩类型，TinyRPC支持四种压缩类型，Raw、Gzip、Snappy、Zlib
	Method       string                  // 方法名
	ID           uint64                  // 请求ID
	RequestLen   uint32                  // 请求体长度
	Checksum     uint32                  // 请求体校验 使用CRC32摘要算法
}

// Marshal will encode request header into a byte slice
func (r *RequestHeader) Marshal() []byte {
	r.RLock()
	defer r.RUnlock()
	idx := 0
	// MaxHeaderSize = 2 + 10 + len(string) + 10 + 10 + 4
	header := make([]byte, MaxHeaderSize+len(r.Method))

	// 将 uint16 数字编码写入 header
	binary.LittleEndian.PutUint16(header[idx:], uint16(r.CompressType))
	idx += Uint16Size

	idx += writeString(header[idx:], r.Method)
	idx += binary.PutUvarint(header[idx:], r.ID)
	idx += binary.PutUvarint(header[idx:], uint64(r.RequestLen))

	binary.LittleEndian.PutUint32(header[idx:], r.Checksum)
	idx += Uint32Size
	return header[:idx]
}

// Unmarshal will decode request header into a byte slice
func (r *RequestHeader) Unmarshal(data []byte) (err error) {
	r.Lock()
	defer r.Unlock()
	if len(data) == 0 {
		return ErrUnmarshal
	}
	defer func() {
		if r := recover(); r != nil {
			err = ErrUnmarshal
		}
	}()

	idx, size := 0, 0
	r.CompressType = compressor.CompressType(binary.LittleEndian.Uint16(data[idx:]))
	idx += Uint16Size

	r.Method, size = readString(data[idx:])
	idx += size

	r.ID, size = binary.Uvarint(data[idx:])
	idx += size

	length, size := binary.Uvarint(data[idx:])
	r.RequestLen = uint32(length)
	idx += size

	r.Checksum = binary.LittleEndian.Uint32(data[idx:])
	return
}

// GetCompressType get compress type
func (r *RequestHeader) GetCompressType() compressor.CompressType {
	r.RLock()
	defer r.RUnlock()
	return r.CompressType
}

// GetMethod get method
func (r *RequestHeader) GetMethod() string {
	r.RLock()
	defer r.RUnlock()
	return r.Method
}

// ResetHeader reset request header
func (r *RequestHeader) ResetHeader() {
	r.Lock()
	defer r.Unlock()
	r.ID = 0
	r.Method = ""
	r.Checksum = 0
	r.CompressType = 0
	r.RequestLen = 0
}

// ResponseHeader request header structure looks like:
// 	+--------------+---------+----------------+-------------+----------+
// 	| CompressType |    ID   |      Error     | ResponseLen | Checksum |
// 	+--------------+---------+----------------+-------------+----------+
// 	|    uint16    | uvarint | uvarint+string |    uvarint  |  uint32  |
// 	+--------------+---------+----------------+-------------+----------+
type ResponseHeader struct {
	sync.RWMutex
	CompressType compressor.CompressType // 压缩类型
	ID           uint64                  // 响应ID号
	Error        string                  // 错误信息
	ResponseLen  uint32                  // 响应体长度
	Checksum     uint32                  // 响应体校验码
}

// Marshal will encode request header into a byte slice
func (r *ResponseHeader) Marshal() []byte {
	r.RLock()
	defer r.RUnlock()
	idx := 0
	// MaxHeaderSize = 2 + 10 + len(string) + 10 + 10 + 4
	header := make([]byte, MaxHeaderSize+len(r.Error))

	// 将 uint16 数字编码写入 header
	binary.LittleEndian.PutUint16(header[idx:], uint16(r.CompressType))
	idx += Uint16Size

	idx += binary.PutUvarint(header[idx:], r.ID)
	idx += writeString(header[idx:], r.Error)
	idx += binary.PutUvarint(header[idx:], uint64(r.ResponseLen))

	binary.LittleEndian.PutUint32(header[idx:], r.Checksum)
	idx += Uint32Size
	return header[:idx]
}

// Unmarshal will decode request header into a byte slice
func (r *ResponseHeader) Unmarshal(data []byte) (err error) {
	r.Lock()
	defer r.Unlock()
	if len(data) == 0 {
		return ErrUnmarshal
	}
	defer func() {
		if r := recover(); r != nil {
			err = ErrUnmarshal
		}
	}()

	idx, size := 0, 0
	r.CompressType = compressor.CompressType(binary.LittleEndian.Uint16(data[idx:]))
	idx += Uint16Size

	r.ID, size = binary.Uvarint(data[idx:])
	idx += size

	r.Error, size = readString(data[idx:])
	idx += size

	length, size := binary.Uvarint(data[idx:])
	r.ResponseLen = uint32(length)
	idx += size

	r.Checksum = binary.LittleEndian.Uint32(data[idx:])
	return
}

// GetCompressType get compress type
func (r *ResponseHeader) GetCompressType() compressor.CompressType {
	r.RLock()
	defer r.RUnlock()
	return r.CompressType
}

// ResetHeader reset request header
func (r *ResponseHeader) ResetHeader() {
	r.Lock()
	defer r.Unlock()
	r.ID = 0
	r.Error = ""
	r.Checksum = 0
	r.CompressType = 0
	r.ResponseLen = 0
}

func readString(data []byte) (string, int) {
	idx := 0
	strLen, size := binary.Uvarint(data)
	idx += size
	str := string(data[idx : idx+int(strLen)])
	idx += len(str)
	return str, idx
}

func writeString(data []byte, str string) int {
	idx := 0
	idx += binary.PutUvarint(data, uint64(len(str)))
	copy(data[idx:], str)
	idx += len(str)
	return idx
}
