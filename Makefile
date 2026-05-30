.PHONY: all build test run install clean

all: build

build:
	go build -o ee .

test:
	go test ./...

run:
	go run .

install:
	go build -o "$(HOME)/.local/bin/ee"

clean:
	rm -f ee
