.PHONY: all build test vet

all: build test vet

build:
	goreleaser check
	goreleaser build --snapshot --clean

test:
	go test -v ./... 

vet:
	go vet ./...

release: build test vet
	git tag -d v1.0.0
	git push origin --delete v1.0.0
	git tag -a v1.0.0 -m "release v1.0.0"
	goreleaser release --clean
