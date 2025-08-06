all: build test
all-deb: build test deb

build:
	go build

deb:
	dpkg-buildpackage --no-sign

deb-arm64:
	dpkg-buildpackage --host-arch arm64 --no-check-builddeps --no-sign

test:
	sudo go test -v ./...

.PHONY: clean

clean:
	rm albius
