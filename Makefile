test:
	go test -v -cover ./...

.PHONY: build
build:
	go build -o build/crawler cmd/app/main.go

exec:
	./build/crawler