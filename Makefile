init:
	mkdir -p ./output/

build: init
	go build -o ./output/bin/ ./...

test: init
	go test -covermode=atomic -race -coverprofile=./output/coverage ./...

smoke:
	go test -tags smoke -covermode=atomic -race -coverprofile=./output/coverage ./...

lint:
	golangci-lint run ./...

fmt:
	go fmt ./...

cover:
	go tool cover -html=./output/coverage

clean:
	rm -rf ./output/
