SHELL = /bin/bash
BIN_NAME = go-tree

.DEFAULT_GOAL=build

build:
	rm -rf build
	env GOOS=linux GOARCH=amd64 go build -o build/${BIN_NAME}-linux-amd64
	env GOOS=darwin GOARCH=amd64 go build -o build/${BIN_NAME}-darwin-amd64
	env GOOS=windows GOARCH=amd64 go build -o build/${BIN_NAME}-win-amd64
