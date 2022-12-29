gen-plugin: # generate protoc-gen-tinyrpc
	go install ./protoc-gen-tinyrpc

proto: # generate go service file
	protoc --go_out=./test_gen --tinyrpc_out=./test_gen ./test_gen/*.proto