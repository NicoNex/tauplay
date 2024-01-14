.PHONY: all

all:
	GOOS=js GOARCH=wasm go build -o assets/play.wasm play.go
	go build
