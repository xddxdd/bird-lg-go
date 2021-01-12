FROM golang:buster AS step_0
#
# IMAGE_ARCH is the binary format of the final output
# BUILD_ARCH is the binaary format of the build host
#
ARG IMAGE_ARCH=amd64
ARG BUILD_ARCH=amd64
#
ENV CGO_ENABLED=0 GOOS=linux GOARCH=$IMAGE_ARCH GO111MODULE=on
WORKDIR /root
COPY . .
# go-bindata is run on the build host as part of the go generate step
RUN GOARCH=$BUILD_ARCH go get -u github.com/kevinburke/go-bindata/...
RUN go generate
RUN go build -ldflags "-w -s" -o /frontend

FROM scratch AS step_1
COPY --from=step_0 /frontend /
ENTRYPOINT ["/frontend"]
