.PHONY: test
test:
	go test -v -cover ./...

test-network:
	go test -tags acceptance_network -run TestCLI_RealSite -v ./test/acceptance/...

.PHONY: build
build:
	go build -o build/crawler cmd/app/main.go

exec:
	./build/crawler --urls https://example.com