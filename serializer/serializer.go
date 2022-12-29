package serializer

// Serializer 对函数传递参数进行序列化和反序列化
type Serializer interface {
	Marshal(message interface{}) ([]byte, error)
	Unmarshal(data []byte, message interface{}) error
}
