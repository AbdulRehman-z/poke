build:
	@go build -o bin/poke main.go

run: build 
	@./bin/poke

test:
	go test -v ./...
