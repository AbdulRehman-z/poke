build:
	@go build -o bin/poke main.go

run: build 
	@clear && ./bin/poke

test:
	go test -v ./...
