BINARY_NAME=dutils
VERSION=1.0.0
LDFLAGS=-ldflags "-X github.com/lukasmay/dutils/cmd.Version=$(VERSION)"

.PHONY: all build install clean test

all: build

build:
	go build $(LDFLAGS) -o $(BINARY_NAME) main.go

install:
	install -d $(DESTDIR)/usr/local/bin
	install -m 755 $(BINARY_NAME) $(DESTDIR)/usr/local/bin/$(BINARY_NAME)

clean:
	rm -f $(BINARY_NAME)

test:
	go test ./...

dev: build
	@echo "Installing $(BINARY_NAME) for local development..."
	@mkdir -p $$HOME/go/bin
	@cp $(BINARY_NAME) $$HOME/go/bin/$(BINARY_NAME)
	@echo "Installed to $$HOME/go/bin/$(BINARY_NAME). Make sure $$HOME/go/bin is in your PATH."
