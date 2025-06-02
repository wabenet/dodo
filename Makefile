.PHONY: all
all: clean test build lint

.PHONY: clean
clean:
	rm -rf ./dist c.out coverage.txt

.PHONY: fmt
fmt:
	go fmt ./...

.PHONY: copier-update
copier-update:
	copier update --trust --skip-answered

.PHONY: update
update:
	go list -f '{{if not (or .Main .Indirect)}}{{.Path}}{{end}}' -m all | xargs --no-run-if-empty go get
	go mod tidy

.PHONY: lint
lint:
	golangci-lint run

.PHONY: test
test:
	go test -race -cover ./...

.PHONY: coverage-report
coverage-report:
	cc-test-reporter before-build
	go test -race -coverprofile=coverage.txt -covermode=atomic -coverpkg=./... ./...
	cp coverage.txt c.out
	cc-test-reporter after-build -t gocov -p $$(go list -m)

.PHONY: build
build:
	goreleaser build --snapshot --clean

.PHONY: release
release:
	goreleaser release --clean
