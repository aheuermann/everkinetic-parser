all: godep dist

dist: build-dist
	 godep go install -v ./...

build-dist:
	godep go build -v ./...

godep:
	go get github.com/tools/godep