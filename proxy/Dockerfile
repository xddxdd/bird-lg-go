FROM golang:buster AS step_0

ENV CGO_ENABLED=0 GO111MODULE=on
WORKDIR /root
COPY . .
RUN go build -ldflags "-w -s" -o /proxy

################################################################################

FROM alpine:edge AS step_1

WORKDIR /root
RUN apk add --no-cache build-base linux-headers
RUN wget https://sourceforge.net/projects/traceroute/files/traceroute/traceroute-2.1.0/traceroute-2.1.0.tar.gz/download \
    -O traceroute-2.1.0.tar.gz
RUN tar xvf traceroute-2.1.0.tar.gz \
    && cd traceroute-2.1.0 \
    && make -j4 LDFLAGS="-static" \
    && strip /root/traceroute-2.1.0/traceroute/traceroute

################################################################################

FROM scratch AS step_2
ENV PATH=/
COPY --from=step_0 /proxy /
COPY --from=step_1 /root/traceroute-2.1.0/traceroute/traceroute /
ENTRYPOINT ["/proxy"]
