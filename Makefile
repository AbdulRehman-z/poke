build:
	@go build -o bin/poke

run: build 
	@./bin/poke

test:
	go test -v ./...
