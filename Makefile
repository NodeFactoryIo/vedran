COMMIT?=${BUILDCOMMIT}
VERSION?=${BUILDTAG}

# enable cgo because it's required by OSX keychain library
CGO_ENABLED=0

# enable go modules
GO111MODULE=on

export CGO_ENABLED
export GO111MODULE

dep:
	go get ./...

test:
	go test ./... -cover

lint:
	golangci-lint run

clean:
	rm vedran 2> /dev/null || exit 0


install:
	make clean
	make build
	cp vedran /usr/local/bin

PLATFORMS := linux/amd64 windows/amd64 darwin/amd64 linux/arm

temp = $(subst /, ,$@)
os = $(word 1, $(temp))
arch = $(word 2, $(temp))
version = $(shell sed -n 's/version=//p' .version)

$(PLATFORMS):
	@if [ "$(os)" = "windows" ]; then \
			GOOS=$(os) GOARCH=$(arch) go build -ldflags "-X github.com/NodeFactoryIo/vedran/pkg/version.Version=$(version)" -o 'build/windows/vedran.exe'; \
	else \
			GOOS=$(os) GOARCH=$(arch) go build -ldflags "-X github.com/NodeFactoryIo/vedran/pkg/version.Version=$(version)" -o 'build/${os}-${arch}/vedran'; \
	fi

buildAll: $(PLATFORMS)

build:
	go build -ldflags "-X github.com/NodeFactoryIo/vedran/pkg/version.Version=$(version)"
