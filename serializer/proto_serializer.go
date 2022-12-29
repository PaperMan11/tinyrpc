package serializer

import (
	"errors"

	"google.golang.org/protobuf/proto"
)

// NotImplementProtoMessageError refers to param not implemented by proto.Message
var ErrNotImplementProtoMessage = errors.New("param does not implement proto.Message")

type ProtoSerializer struct{}

func NewProtoSerializer() Serializer {
	return &ProtoSerializer{}
}

func (*ProtoSerializer) Marshal(message interface{}) ([]byte, error) {
	var body proto.Message
	if message == nil {
		return []byte{}, nil
	}
	var ok bool
	if body, ok = message.(proto.Message); !ok {
		return nil, ErrNotImplementProtoMessage
	}
	return proto.Marshal(body)
}

func (*ProtoSerializer) Unmarshal(data []byte, message interface{}) error {
	var body proto.Message
	if message == nil {
		return nil
	}

	var ok bool
	body, ok = message.(proto.Message)
	if !ok {
		return ErrNotImplementProtoMessage
	}

	return proto.Unmarshal(data, body)
}
