all: build

.PHONY: clean
clean:
	rm dodo_*

.PHONY: fmt
fmt:
	go fmt ./...

.PHONY: lint
lint:
	gometalinter ./...

.PHONY: test
test:
	go test ./...

.PHONY: dep
dep:
	dep ensure

.PHONY: build
build:
	gox -arch="amd64" -os="darwin" ./...
