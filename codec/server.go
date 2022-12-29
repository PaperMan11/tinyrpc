package codec

import (
	"bufio"
	"hash/crc32"
	"io"
	"net/rpc"
	"sync"
	"tinyrpc/compressor"
	"tinyrpc/header"
	"tinyrpc/serializer"
)

type reqCtx struct {
	requestID   uint64
	compareType compressor.CompressType
}

type serverCodec struct {
	r io.Reader
	w io.Writer
	c io.Closer

	request    header.RequestHeader
	serializer serializer.Serializer
	mu         sync.Mutex
	seq        uint64
	pending    map[uint64]*reqCtx
}

// NewServerCodec Create a new server codec
func NewServerCodec(conn io.ReadWriteCloser, serializer serializer.Serializer) rpc.ServerCodec {
	return &serverCodec{
		r:          bufio.NewReader(conn),
		w:          bufio.NewWriter(conn),
		c:          conn,
		serializer: serializer,
		pending:    make(map[uint64]*reqCtx),
	}
}

// ReadRequestHeader read the rpc request header from the io stream
func (s *serverCodec) ReadRequestHeader(r *rpc.Request) error {
	s.request.ResetHeader()
	data, err := recvFrame(s.r)
	if err != nil {
		return nil
	}
	err = s.request.Unmarshal(data)
	if err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.seq++
	s.pending[s.seq] = &reqCtx{
		requestID:   s.request.ID,
		compareType: s.request.GetCompressType(),
	}
	r.ServiceMethod = s.request.GetMethod()
	r.Seq = s.seq // response 时会用到
	return nil
}

// ReadRequestBody read the rpc request body from the io stream
func (s *serverCodec) ReadRequestBody(param interface{}) error {
	if param == nil {
		if s.request.RequestLen != 0 {
			// 也需要读出来
			if err := read(s.r, make([]byte, int(s.request.RequestLen))); err != nil {
				return err
			}
		}
		return nil
	}

	reqBody := make([]byte, int(s.request.RequestLen))
	err := read(s.r, reqBody)
	if err != nil {
		return err
	}

	if s.request.Checksum != 0 {
		if crc32.ChecksumIEEE(reqBody) != s.request.Checksum {
			return ErrUnexpectedChecksum
		}
	}

	c, ok := compressor.Compressors[s.request.GetCompressType()]
	if !ok {
		return ErrNotFoundCompressor
	}

	req, err := c.Unzip(reqBody) // 解压缩
	if err != nil {
		return err
	}

	return s.serializer.Unmarshal(req, param) // 反序列化
}

// WriteResponse Write the rpc response header and body to the io stream
func (s *serverCodec) WriteResponse(resp *rpc.Response, param interface{}) error {
	s.mu.Lock()
	reqCtx, ok := s.pending[resp.Seq]
	if !ok {
		s.mu.Unlock()
		return ErrInvalidSequence
	}
	delete(s.pending, resp.Seq)
	s.mu.Unlock()

	if resp.Error != "" { // 如果RPC调用结果有误，把param置为nil
		param = nil
	}

	c, ok := compressor.Compressors[reqCtx.compareType]
	if !ok {
		return ErrNotFoundCompressor
	}

	var (
		respBody           []byte // marshal
		compressedRespBody []byte // zip
		err                error
	)
	if param != nil {
		respBody, err = s.serializer.Marshal(param) // 序列化
		if err != nil {
			return err
		}
	}

	compressedRespBody, err = c.Zip(respBody) // 压缩
	if err != nil {
		return err
	}
	h := header.ResponsePool.Get().(*header.ResponseHeader)
	defer func() {
		h.ResetHeader()
		header.ResponsePool.Put(h)
	}()
	h.ID = reqCtx.requestID
	h.Error = resp.Error
	h.ResponseLen = uint32(len(compressedRespBody))
	h.Checksum = crc32.ChecksumIEEE(compressedRespBody)
	h.CompressType = reqCtx.compareType
	// 发送响应头
	if err = sendFrame(s.w, h.Marshal()); err != nil {
		return err
	}
	// 发送响应体
	if err = write(s.w, compressedRespBody); err != nil {
		return err
	}
	s.w.(*bufio.Writer).Flush()
	return nil
}

// Close can be called multiple times and must be idempotent.
func (s *serverCodec) Close() error {
	return s.c.Close()
}
