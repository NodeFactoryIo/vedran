COMMIT?=${BUILDCOMMIT}
VERSION?=${BUILDTAG}

# enable cgo because it's required by OSX keychain library
CGO_ENABLED=1

# enable go modules
GO111MODULE=on

export CGO_ENABLED
export GO111MODULE

dep:
	go get ./...

test:
	go test ./...

lint:
	golangci-lint run

clean:
	rm vedran-daemon 2> /dev/null || exit 0

build:
	go build

install:
	make clean
	make build
	cp vedran-daemon /usr/local/bin
