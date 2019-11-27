all: clean test build

.PHONY: clean
clean:
	rm -f -r bin/
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
test: proto/stage.pb.go
	go generate ./...
	go test -cover ./...

.PHONY: build
build: clean build/cmd build/plugins

.PHONY: build/cmd
build/cmd: proto/stage.pb.go
	go generate ./...
	gox -arch="amd64" -os="darwin linux" -output "./bin/{{.Dir}}_{{.OS}}_{{.Arch}}" ./cmd/...

.PHONY: build/plugins
build/plugins: proto/stage.pb.go
	go generate ./...
	gox -arch="amd64" -os="darwin linux" -output "./bin/plugins/{{.Dir}}_{{.OS}}_{{.Arch}}" ./plugins/...

proto/%.pb.go: proto/%.proto
	protoc --go_out=plugins=grpc:. $<
