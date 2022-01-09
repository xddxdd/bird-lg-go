FROM golang:buster AS step_0
ENV CGO_ENABLED=0 GO111MODULE=on
WORKDIR /root
COPY . .
RUN go build -ldflags "-w -s" -o /frontend

################################################################################

FROM scratch AS step_1
COPY --from=step_0 /frontend /
ENTRYPOINT ["/frontend"]
