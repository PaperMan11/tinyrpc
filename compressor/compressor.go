package compressor

type CompressType uint16

const (
	Raw CompressType = iota
	Gzip
	Snappy
	Zlib
)

var Compressors = map[CompressType]Compressor{}

// Compressor 对函数传递参数进行压缩和解压缩
type Compressor interface {
	Zip([]byte) ([]byte, error)
	Unzip([]byte) ([]byte, error)
}
