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
RUN go build -ldflags "-w -s" -o /proxy

################################################################################

#if defined(ARCH_AMD64)
FROM amd64/debian:sid AS step_1
ENV TARGET_ARCH=x86_64
#elif defined(ARCH_I386)
FROM i386/debian:sid AS step_1
ENV TARGET_ARCH=i386
#elif defined(ARCH_ARM32V7)
FROM arm32v7/debian:sid AS step_1
ENV TARGET_ARCH=arm
#elif defined(ARCH_ARM64V8)
FROM arm64v8/debian:sid AS step_1
ENV TARGET_ARCH=aarch64
#elif defined(ARCH_PPC64LE)
FROM ppc64le/debian:sid AS step_1
ENV TARGET_ARCH=ppc64le
#elif defined(ARCH_S390X)
FROM s390x/debian:sid AS step_1
ENV TARGET_ARCH=s390
#else
#error "Architecture not set"
#endif

WORKDIR /root
RUN apt-get -qq update && DEBIAN_FRONTEND=noninteractive apt-get -qq install -y \
        build-essential musl-dev musl-tools tar wget git
RUN git clone https://github.com/sabotage-linux/kernel-headers.git
RUN wget https://sourceforge.net/projects/traceroute/files/traceroute/traceroute-2.1.0/traceroute-2.1.0.tar.gz/download \
        -O traceroute-2.1.0.tar.gz
RUN tar xvf traceroute-2.1.0.tar.gz \
    && cd traceroute-2.1.0 \
    && make -j4 CC=musl-gcc CFLAGS="-I/root/kernel-headers/${TARGET_ARCH}/include" LDFLAGS="-static"

################################################################################

FROM scratch AS step_2
ENV PATH=/
COPY --from=step_0 /proxy /
COPY --from=step_1 /root/traceroute-2.1.0/traceroute/traceroute /
ENTRYPOINT ["/proxy"]
