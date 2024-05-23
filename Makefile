.PHONY: all build test vet

all: build test vet

build:
	goreleaser check
	goreleaser build --snapshot --clean

test:
	go test -v ./... 

vet:
	go vet ./...

release: test vet		
	goreleaser check
	goreleaser release --clean

tagging:
	git push origin --delete v1.0.0
	git tag -d v1.0.0
	git tag -a v1.0.0 -m "release v1.0.0"