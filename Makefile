.PHONY: dirs
dirs:
	mkdir -p pkg/ruby/versions
	mkdir -p pkg/python/versions

.PHONY: build
build: dirs
	go build ./...

.PHONY: clean
clean:
	rm -rf target

.PHONY: test
test: test
	go test ./...

.PHONY: test/integration
test/integration:
	go test -tags=integration ./tests/integration/...

.PHONY: test/integration/update
test/integration/update:
	go test -count=1 -race -tags=integration ./tests/integration/... -update

.PHONY: check
check: vet lint

.PHONY: vet
vet:
	go vet ./...

.PHONY: lint
lint:
	go run github.com/golangci/golangci-lint/cmd/golangci-lint@latest run

.PHONY: fix
fix:
	go run github.com/golangci/golangci-lint/cmd/golangci-lint@latest run --fix

.PHONY: format
format: tagalign
	go run mvdan.cc/gofumpt@latest -l -w .

.PHONY: tagalign
tagalign:
	go run github.com/4meepo/tagalign/cmd/tagalign@latest -fix -sort ./...
