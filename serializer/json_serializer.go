package serializer

import "encoding/json"

type Json struct{}

func NewJsonSerializer() Serializer {
	return &Json{}
}

// Marshal .
func (*Json) Marshal(message interface{}) ([]byte, error) {
	return json.Marshal(message)
}

// Unmarshal .
func (*Json) Unmarshal(data []byte, message interface{}) error {
	return json.Unmarshal(data, message)
}
