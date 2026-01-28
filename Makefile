.PHONY: test build clean

build:
	go build -o fm ./cmd/fm

test:
	go test ./...

clean:
	rm -f fm
