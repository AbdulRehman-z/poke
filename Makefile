run: build
	@./bin/poke

build:
	@go build -o bin/poke main.go

test:
	@go test -v ./... -count=1
