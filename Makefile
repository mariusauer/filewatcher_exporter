GOARCH ?= amd64
BINARY := bin/filewatcher-exporter

.PHONY: build-linux clean

build-linux:
	@mkdir -p $(dir $(BINARY))
	GOOS=linux GOARCH=$(GOARCH) go build -o $(BINARY) .

clean:
	rm -f $(BINARY)
