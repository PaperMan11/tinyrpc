## 1 生成自定义服务文件
`protoc-gen-tinyrpc` 为自定义插件，可帮助开发者将 proto 文件生成 tinyrpc 对应的服务文件（.srv.go）和 proto 序列化与反序列化文件（.pb.go）。

> 自定义插件编写推荐阅读：https://mdnice.com/writing/ab4aec3d6936437f904cd18c1996ce4e
> 源码：https://github.com/yusank/protoc-gen-go-http


## 2 TinyRpc 
&emsp;&emsp;TinyRpc 是基于 Go 语言标准库 net/rpc 扩展的远程过程调用框架，它具有以下特性：
- 基于 TCP 传输层协议支持多种压缩格式：gzip、snappy、zlib；
- 基于二进制的 Protocol Buffer 序列化协议：具有协议编码小及高扩展性和跨平台性；
- 支持自定义序列化器。
- 支持生成工具：TinyRPC提供的 protoc-gen-tinyrpc 插件可以帮助开发者快速定义自己的服务；

## 3 标准库 net/rpc 网络框架

### 3.1 Server
**rpc.Server：** 
```go
// Server represents an RPC Server.
type Server struct {
	serviceMap sync.Map   // map[string]*service
	reqLock    sync.Mutex // protects freeReq
	freeReq    *Request   // 相当于 Request 对象池
	respLock   sync.Mutex // protects freeResp
	freeResp   *Response  // 相当于 Response 对象池
}

type Request struct {
	ServiceMethod string   // format: "Service.Method"
	Seq           uint64   // sequence number chosen by client
	next          *Request // for free list in Server
}

type Response struct {
	ServiceMethod string    // echoes that of the Request
	Seq           uint64    // echoes that of the request
	Error         string    // error, if any.
	next          *Response // for free list in Server
}
```

**rpc.service：** 注册服务对象模型
```go
type methodType struct {
	sync.Mutex // protects counters
	method     reflect.Method
	ArgType    reflect.Type
	ReplyType  reflect.Type
	numCalls   uint
}

type service struct {
	name   string                 // name of service
	rcvr   reflect.Value          // receiver of methods for the service
	typ    reflect.Type           // type of the receiver
	method map[string]*methodType // registered methods
}
```

**服务注册流程：**
```go
1. 通过调用 rpc.NewServer() 创建 Server 实例。
2. Server.Register() 或 Server.RegisterName() 将服务注册入 Server.serviceMap 中。
```

**rpc Server 执行流程：**
```mermaid
graph
    listen("net.Listen()")
    accept("lis.Accept()")
    serveConn("server.ServeConn()")
```
