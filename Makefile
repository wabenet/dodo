all: clean test build

.PHONY: clean
clean:
	rm -f dodo_*

.PHONY: fmt
fmt:
	go fmt ./...

.PHONY: tidy
tidy:
	go mod tidy

.PHONY: lint
lint:
	golangci-lint run --enable-all

.PHONY: test
test:
	go test -cover ./...

.PHONY: vendor
vendor: Gopkg.lock Gopkg.toml
	dep ensure

.PHONY: build
build: clean
	gox -arch="amd64" -os="darwin linux" ./...
