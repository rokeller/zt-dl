VERSION_RAW = $(shell git describe --tags 2>/dev/null || echo '0.0.0')

zt-dl: $(wildcard **/*.go)
	@go build -ldflags \
		"-X github.com/rokeller/zt-dl/cmd.version=${VERSION_RAW}-local"

.PHONY: release
release:
	@CGO_ENABLED=0 go build -ldflags "-s -w \
		-X github.com/rokeller/zt-dl/cmd.version=${VERSION_RAW}-local"

.PHONY: client
client:
	@pnpm -C server/client/ build

.PHONY: test
test: zt-dl
	@go test ./...

.PHONY: cover
cover: zt-dl
	@go test ./... -coverprofile=coverage.full.out
	@cat coverage.full.out | grep -v "test/" > coverage.out
	@go tool cover -func=coverage.out
	@go tool cover -html=coverage.out -o coverage.html
	@go-cover-treemap -coverprofile coverage.out > coverage.svg

.PHONY: tidy
tidy:
	@go mod tidy

.PHONY: update
update:
	@go get -u ./...

.PHONY: clean
clean:
	@rm -rf zt-dl
