package serializer

import (
	"testing"
	pb "tinyrpc/test_gen/message"

	"github.com/stretchr/testify/assert"
)

type testArg struct{}

func TestProtoSerializer_Marshal(t *testing.T) {
	type expect struct {
		data []byte
		err  error
	}
	cases := []struct {
		name   string
		arg    interface{}
		expect expect
	}{
		{
			name: "test-1",
			arg:  &pb.ArithRequest{A: 1, B: 2},
			expect: expect{
				data: []byte{0x9, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0xf0,
					0x3f, 0x11, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x40},
				err: nil,
			},
		},
		{
			name: "test-2",
			arg:  testArg{},
			expect: expect{
				data: nil,
				err:  ErrNotImplementProtoMessage,
			},
		},
		{
			name: "test-3",
			arg:  nil,
			expect: expect{
				data: []byte{},
				err:  nil,
			},
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			data, err := NewProtoSerializer().Marshal(c.arg)
			assert.Equal(t, c.expect.data, data)
			assert.Equal(t, c.expect.err, err)
		})
	}
}

func TestProtoSerializer_Unmarshal(t *testing.T) {
	type expect struct {
		message interface{}
		err     error
	}
	cases := []struct {
		name    string
		arg     []byte
		message interface{}
		expect  expect
	}{
		{
			name: "test-1",
			arg: []byte{0x9, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0xf0,
				0x3f, 0x11, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x40},
			message: &pb.ArithRequest{},
			expect: expect{
				message: &pb.ArithRequest{A: 1, B: 2},
				err:     nil,
			},
		},
		{
			name:    "test-2",
			arg:     nil,
			message: nil,
			expect: expect{
				message: nil,
				err:     nil,
			},
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			err := NewProtoSerializer().Unmarshal(c.arg, c.message)
			if err != nil {
				assert.Equal(t, c.expect.message.(*pb.ArithRequest).A,
					c.message.(*pb.ArithRequest).A)
				assert.Equal(t, c.expect.message.(*pb.ArithRequest).B,
					c.message.(*pb.ArithRequest).B)
			}

			assert.Equal(t, c.expect.err, err)
		})
	}
}
