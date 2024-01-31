build:
	@go build -o bin/chain

run: build
	@./bin/chain

test:
	@go test -v ./...