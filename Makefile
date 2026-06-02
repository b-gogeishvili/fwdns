# Simple automation for building and packaging fwdns.
# Run `make` (or `make build`) to compile, `make test` to run the unit tests.

BINARY := fwdns

.PHONY: build run build-windows docker clean

build:
	go build -o $(BINARY) .

run: build
	./$(BINARY)

build-windows:
	GOOS=windows GOARCH=amd64 go build -o $(BINARY).exe .

docker:
	docker build -t $(BINARY) .

clean:
	rm -f $(BINARY) $(BINARY).exe
