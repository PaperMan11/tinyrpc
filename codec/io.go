package codec

import (
	"encoding/binary"
	"io"
	"net"
)

// sendFrame write requestHeadr or responseHeader
// 若写入数据的长度为 0 ，此时sendFrame 函数会向IO流写入uvarint类型的 0 值；
// 若写入数据的长度大于 0 ，此时sendFrame 函数会向IO流写入uvarint类型的 len(data) 值，随后将该字节串的数据 data 写入IO流中。
func sendFrame(w io.Writer, data []byte) (err error) {
	var size [binary.MaxVarintLen64]byte

	if len(data) == 0 {
		n := binary.PutUvarint(size[:], uint64(0))
		return write(w, size[:n])
	}

	n := binary.PutUvarint(size[:], uint64(len(data)))
	if err = write(w, size[:n]); err != nil {
		return
	}
	return write(w, data)
}

// recvFrame read requestHeadr or responseHeader
// 首先会向IO中读入uvarint类型的 size ，表示要接收数据的长度，
// 随后将该从IO流中读取该 size 长度字节串。
func recvFrame(r io.Reader) (data []byte, err error) {
	size, err := binary.ReadUvarint(r.(io.ByteReader))
	if err != nil {
		return nil, err
	}
	if size != 0 {
		data = make([]byte, size)
		if err = read(r, data); err != nil {
			return nil, err
		}
	}
	return data, nil
}

// write 写入指定 data
func write(w io.Writer, data []byte) error {
	for i := 0; i < len(data); {
		n, err := w.Write(data[i:])
		if _, ok := err.(net.Error); !ok {
			return err
		}
		i += n
	}
	return nil
}

// read 读取内容至 data
func read(r io.Reader, data []byte) error {
	for i := 0; i < len(data); {
		n, err := r.Read(data[i:])
		if err != nil {
			if _, ok := err.(net.Error); !ok {
				return err
			}
		}
		i += n
	}
	return nil
}
