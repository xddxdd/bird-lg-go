.PHONY: all
all: $(shell find . -name \*.go -type f)
	go build -ldflags "-w -s" -o proxy
