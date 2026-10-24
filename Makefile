.PHONY: build clean test fmt

build:
	go build -o bin/gm .

test:
	go test ./...

fmt:
	go fmt ./...

clean:
	rm -rf bin
