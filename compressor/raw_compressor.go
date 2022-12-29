package compressor

func init() {
	Compressors[Raw] = NewRawCompressor()
}

type RawCompressor struct{}

func NewRawCompressor() Compressor {
	return &RawCompressor{}
}

func (*RawCompressor) Zip(data []byte) ([]byte, error) {
	return data, nil
}

func (*RawCompressor) Unzip(data []byte) ([]byte, error) {
	return data, nil
}
