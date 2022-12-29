package tinyrpc

import (
	"errors"
	"log"
	"net"
	"net/rpc"
	"reflect"
	"testing"
	"tinyrpc/compressor"
	"tinyrpc/serializer"
	js "tinyrpc/test_gen/json"
	pb "tinyrpc/test_gen/message"

	"github.com/stretchr/testify/assert"
)

// init Server
func init() {
	// proto serializer
	lis, err := net.Listen("tcp", ":8008")
	if err != nil {
		log.Fatal(err)
	}

	server := NewServer()
	err = server.Register(new(pb.ArithService))
	if err != nil {
		log.Fatal(err)
	}
	go server.Serve(lis)

	// json serializer
	lis, err = net.Listen("tcp", ":8009")
	if err != nil {
		log.Fatal(err)
	}

	server = NewServer(WithSerializer(serializer.NewJsonSerializer()))
	err = server.Register(new(js.ArithService))
	if err != nil {
		log.Fatal(err)
	}
	go server.Serve(lis)
}

func TestServer_Register(t *testing.T) {
	server := NewServer()
	err := server.RegisterName("ArithService", new(pb.ArithService))
	assert.Equal(t, nil, err)
	err = server.Register(new(pb.ArithService))
	assert.Equal(t, errors.New("rpc: service already defined: ArithService"), err)
}

// --------------------------client---------------------------

const (
	Call      = true
	AsyncCall = false
)

func client_call(t *testing.T, comporessType compressor.CompressType, call bool) {
	conn, err := net.Dial("tcp", ":8008")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()
	client := NewClient(conn, WithCompress(comporessType))
	defer client.Close()

	// test
	type expect struct {
		reply *pb.ArithResponse
		err   error
	}
	cases := []struct {
		client         *Client
		name           string
		serviceMenthod string
		arg            *pb.ArithRequest
		expect         expect
	}{
		{
			client:         client,
			name:           "test-1",
			serviceMenthod: "ArithService.Add",
			arg:            &pb.ArithRequest{A: 20, B: 5},
			expect: expect{
				reply: &pb.ArithResponse{C: 25},
				err:   nil,
			},
		},
		{
			client:         client,
			name:           "test-2",
			serviceMenthod: "ArithService.Sub",
			arg:            &pb.ArithRequest{A: 20, B: 5},
			expect: expect{
				reply: &pb.ArithResponse{C: 15},
				err:   nil,
			},
		},
		{
			client:         client,
			name:           "test-3",
			serviceMenthod: "ArithService.Mul",
			arg:            &pb.ArithRequest{A: 20, B: 5},
			expect: expect{
				reply: &pb.ArithResponse{C: 100},
				err:   nil,
			},
		},
		{
			client:         client,
			name:           "test-4",
			serviceMenthod: "ArithService.Div",
			arg:            &pb.ArithRequest{A: 20, B: 5},
			expect: expect{
				reply: &pb.ArithResponse{C: 4},
			},
		},
		{
			client,
			"test-5",
			"ArithService.Div",
			&pb.ArithRequest{A: 20, B: 0},
			expect{
				&pb.ArithResponse{},
				rpc.ServerError("divided is zero"),
			},
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			reply := &pb.ArithResponse{}
			var err error
			if call {
				err = c.client.Call(c.serviceMenthod, c.arg, reply)
			} else {
				c := <-c.client.AsyncCall(c.serviceMenthod, c.arg, reply)
				err = c.Error
			}

			assert.Equal(t, true, reflect.DeepEqual(c.expect.reply.C, reply.C))
			assert.Equal(t, c.expect.err, err)
		})
	}
}

// TestNewClientWithSnappyCompress test raw comressor
func TestNewClientWithRawCompress(t *testing.T) {
	client_call(t, compressor.Raw, Call)
	client_call(t, compressor.Raw, AsyncCall)
}

// TestNewClientWithSnappyCompress test snappy comressor
func TestNewClientWithSnappyCompress(t *testing.T) {
	client_call(t, compressor.Snappy, Call)
	client_call(t, compressor.Snappy, AsyncCall)
}

// TestNewClientWithGzipCompress test gzip comressor
func TestNewClientWithGzipCompress(t *testing.T) {
	client_call(t, compressor.Gzip, Call)
	client_call(t, compressor.Gzip, AsyncCall)
}

// TestNewClientWithZlibCompress test zlib compressor
func TestNewClientWithZlibCompress(t *testing.T) {
	client_call(t, compressor.Zlib, Call)
	client_call(t, compressor.Zlib, AsyncCall)
}

// TestNewClientWithSerializer .
func TestNewClientWithSerializer(t *testing.T) {

	conn, err := net.Dial("tcp", ":8009")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()
	client := NewClient(conn, WithSerializer(serializer.NewJsonSerializer()))
	defer client.Close()

	type expect struct {
		reply *js.ArithResponse
		err   error
	}
	cases := []struct {
		client         *Client
		name           string
		serviceMenthod string
		arg            *js.ArithRequest
		expect         expect
	}{
		{
			client:         client,
			name:           "test-1",
			serviceMenthod: "ArithService.Add",
			arg:            &js.ArithRequest{A: 20, B: 5},
			expect: expect{
				reply: &js.ArithResponse{C: 25},
			},
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			reply := &js.ArithResponse{}
			err := c.client.Call(c.serviceMenthod, c.arg, reply)
			assert.Equal(t, true, reflect.DeepEqual(c.expect.reply.C, reply.C))
			assert.Equal(t, c.expect.err, err)
		})
	}
}
