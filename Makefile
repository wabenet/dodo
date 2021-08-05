all: clean build

.PHONY: clean
clean:
	rm -f main.go
	rm -f dodo_*

.PHONY: tidy
tidy:
	go mod tidy

.PHONY: lint
lint:
	golangci-lint run --enable-all

.PHONY: build
build:
	goreleaser build --snapshot --rm-dist
