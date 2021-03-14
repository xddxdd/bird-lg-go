.PHONY: all
all:
	go get -u github.com/kevinburke/go-bindata/...
	go generate
	go build -ldflags "-w -s" -o frontend
