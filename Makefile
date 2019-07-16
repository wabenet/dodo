all: clean test build

.PHONY: clean
clean:
	rm -f dodo_*
	rm -f virtualbox_*
	rm -f proto/*.pb.go

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

.PHONY: build
build: clean proto/provider.pb.go
	gox -arch="amd64" -os="darwin linux" ./...

proto/%.pb.go: proto/%.proto
	protoc --go_out=plugins=grpc:. $<
