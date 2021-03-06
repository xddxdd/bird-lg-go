FROM golang:buster AS step_0

#if defined(ARCH_AMD64)
ENV GOOS=linux GOARCH=amd64
#elif defined(ARCH_I386)
ENV GOOS=linux GOARCH=386
#elif defined(ARCH_ARM32V7)
ENV GOOS=linux GOARCH=arm
#elif defined(ARCH_ARM64V8)
ENV GOOS=linux GOARCH=arm64
#elif defined(ARCH_PPC64LE)
ENV GOOS=linux GOARCH=ppc64le
#elif defined(ARCH_S390X)
ENV GOOS=linux GOARCH=s390x
#else
#error "Architecture not set"
#endif

ENV CGO_ENABLED=0 GO111MODULE=on
WORKDIR /root
COPY . .
# go-bindata is run on the build host as part of the go generate step
RUN GOARCH=amd64 go get -u github.com/kevinburke/go-bindata/...
RUN go generate
RUN go build -ldflags "-w -s" -o /frontend

################################################################################

FROM scratch AS step_1
COPY --from=step_0 /frontend /
ENTRYPOINT ["/frontend"]
