FROM golang AS step_0
ENV CGO_ENABLED=0 GO111MODULE=on
WORKDIR /root
COPY . .
RUN go build -ldflags "-w -s" -o /frontend

################################################################################

FROM alpine:edge AS step_1

WORKDIR /root
RUN apk add --no-cache build-base pkgconf perl gettext

RUN wget https://github.com/rfc1036/whois/archive/refs/tags/v5.5.18.tar.gz \
    -O whois-5.5.18.tar.gz
RUN tar xvf whois-5.5.18.tar.gz \
    && cd whois-5.5.18 \
    && make -j4 LDFLAGS="-static" \
    && strip /root/whois-5.5.18/whois

################################################################################

FROM scratch AS step_2
ENV PATH=/
ENV BIRDLG_WHOIS=/whois
COPY --from=step_0 /frontend /
COPY --from=step_1 /root/whois-5.5.18/whois /
ENTRYPOINT ["/frontend"]
