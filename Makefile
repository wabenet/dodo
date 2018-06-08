all: clean lint test build

.PHONY: clean
clean:
	rm -f dodo_*

.PHONY: fmt
fmt:
	go fmt ./...

.PHONY: lint
lint:
	gometalinter ./...

.PHONY: test
test:
	go test -cover ./...

.PHONY: dep
dep:
	dep ensure

.PHONY: build
build: clean
	gox -arch="amd64" -os="darwin" ./...
