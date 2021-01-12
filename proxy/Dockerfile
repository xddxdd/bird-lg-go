FROM golang:buster AS step_0
#
# IMAGE_ARCH is the binary format of the final output
#
ARG IMAGE_ARCH=amd64
#
ENV CGO_ENABLED=0 GOOS=linux GOARCH=$IMAGE_ARCH GO111MODULE=on
WORKDIR /root
COPY . .
RUN go build -ldflags "-w -s" -o /proxy

FROM amd64/debian AS step_1
ENV TARGET_ARCH=x86_64
WORKDIR /root
RUN apt-get -qq update && DEBIAN_FRONTEND=noninteractive apt-get -qq install -y \
        build-essential musl-dev musl-tools tar wget git
RUN git clone https://github.com/sabotage-linux/kernel-headers.git
RUN wget https://sourceforge.net/projects/traceroute/files/traceroute/traceroute-2.1.0/traceroute-2.1.0.tar.gz/download \
        -O traceroute-2.1.0.tar.gz
RUN tar xvf traceroute-2.1.0.tar.gz \
    && cd traceroute-2.1.0 \
    && make -j4 CC=musl-gcc CFLAGS="-I/root/kernel-headers/${TARGET_ARCH}/include" LDFLAGS="-static"

FROM scratch AS step_2
ENV PATH=/
COPY --from=step_0 /proxy /
COPY --from=step_1 /root/traceroute-2.1.0/traceroute/traceroute /
ENTRYPOINT ["/proxy"]
