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

type clientCodec struct {
	r io.Reader
	w io.Writer
	c io.Closer

	compressor compressor.CompressType // rpc compress type(raw,gzip,snappy,zlib)
	serializer serializer.Serializer
	response   header.ResponseHeader // rpc response header
	mu         sync.Mutex            // protect pending map
	pending    map[uint64]string
}

// NewClientCodec Create a new client codec
func NewClientCodec(conn io.ReadWriteCloser,
	compressType compressor.CompressType,
	serializer serializer.Serializer) rpc.ClientCodec {
	return &clientCodec{
		r:          bufio.NewReader(conn),
		w:          bufio.NewWriter(conn),
		c:          conn,
		compressor: compressType,
		serializer: serializer,
		pending:    make(map[uint64]string),
	}
}

// WriteRequest Write the rpc request header and body to the io stream
func (c *clientCodec) WriteRequest(r *rpc.Request, param interface{}) error {
	c.mu.Lock()
	c.pending[r.Seq] = r.ServiceMethod
	c.mu.Unlock()

	cpr, ok := compressor.Compressors[c.compressor]
	if !ok {
		return ErrNotFoundCompressor
	}
	reqBody, err := c.serializer.Marshal(param)
	if err != nil {
		return err
	}
	compressedReqBody, err := cpr.Zip(reqBody)
	if err != nil {
		return err
	}
	h := header.RequestPool.Get().(*header.RequestHeader)
	defer func() {
		h.ResetHeader()
		header.RequestPool.Put(h)
	}()
	h.ID = r.Seq
	h.Method = r.ServiceMethod
	h.RequestLen = uint32(len(compressedReqBody))
	h.CompressType = c.compressor
	h.Checksum = crc32.ChecksumIEEE(compressedReqBody)

	if err := sendFrame(c.w, h.Marshal()); err != nil {
		return err
	}

	if err := write(c.w, compressedReqBody); err != nil {
		return err
	}
	c.w.(*bufio.Writer).Flush()
	return nil
}

// ReadResponseHeader read the rpc response header from the io stream
func (c *clientCodec) ReadResponseHeader(resp *rpc.Response) error {
	c.response.ResetHeader()
	data, err := recvFrame(c.r)
	if err != nil {
		return err
	}
	if err = c.response.Unmarshal(data); err != nil {
		return err
	}

	c.mu.Lock()
	defer c.mu.Unlock()
	resp.Seq = c.response.ID
	resp.Error = c.response.Error
	resp.ServiceMethod = c.pending[resp.Seq]
	delete(c.pending, resp.Seq)
	return nil
}

// ReadResponseBody read the rpc response body from the io stream
func (c *clientCodec) ReadResponseBody(param interface{}) error {
	if param == nil {
		if c.response.ResponseLen != 0 {
			return read(c.r, make([]byte, int(c.response.ResponseLen)))
		}
		return nil
	}
	respBody := make([]byte, int(c.response.ResponseLen))
	if err := read(c.r, respBody); err != nil {
		return err
	}

	if c.response.Checksum != 0 {
		if crc32.ChecksumIEEE(respBody) != c.response.Checksum {
			return ErrUnexpectedChecksum
		}
	}

	if c.response.GetCompressType() != c.compressor {
		return ErrCompressorTypeMismatch
	}

	resp, err := compressor.Compressors[c.compressor].Unzip(respBody)
	if err != nil {
		return err
	}
	return c.serializer.Unmarshal(resp, param)
}

func (c *clientCodec) Close() error {
	return c.c.Close()
}
