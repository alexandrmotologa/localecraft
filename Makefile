.PHONY: all build build-web build-cli test audit clean

BINARY_NAME=localecraft

all: build

build-web:
	cd web && npm install && npm run build

build-cli:
	go build -o $(BINARY_NAME) ./cmd/localecraft

build: build-web build-cli

test:
	go test -v ./...

audit:
	./$(BINARY_NAME) audit --config ./examples/localecraft.yaml

clean:
	rm -f $(BINARY_NAME) $(BINARY_NAME).exe
	rm -rf web/dist
