.PHONY: install gen test

install:
	go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.36.11
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.6.2

gen: 
	protoc --version
	protoc-gen-go --version
	protoc-gen-go-grpc --version
	protoc --go_out=. --go-grpc_out=. *.proto
	protoc --go_out=. --go-grpc_out=. *.proto

test:
	go test -v -race