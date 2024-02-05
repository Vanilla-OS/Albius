all: build test
all-deb: build test deb

build:
	go build

deb:
	dpkg-buildpackage --no-sign

test:
	sudo go test -v ./...

.PHONY: clean

clean:
	rm albius
