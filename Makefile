build:
	@go build -o bin/poke main.go

run: build 
	@clear && ./bin/poke

protocode:
	protoc --go_out=. --go_opt=paths=source_relative \
    --go-grpc_out=. --go-grpc_opt=paths=source_relative \
    proto/service.proto	

test:
	go test -v ./...

