GOARCH ?= amd64
BINARY := bin/filewatcher_exporter

.PHONY: build clean

build:
	@mkdir -p $(dir $(BINARY))
	CGO_ENABLED=0 GOOS=linux GOARCH=$(GOARCH) go build -o $(BINARY) .

clean:
	rm -f $(BINARY)
