all: clean build

.PHONY: clean
clean:
	rm -f main.go
	rm -f dodo_*

.PHONY: lint
lint:
	golangci-lint run --enable-all

.PHONY: build
build:
	go generate .
	gox -arch="amd64" -os="darwin linux" .
