ARG THIS_ARCH_ALT=amd64
FROM multiarch/alpine:${THIS_ARCH_ALT}-edge

LABEL Lan Tian "lantian@lantian.pub"
ENV GOOS=linux GOARCH=${THIS_ARCH_GO}
WORKDIR /root
COPY . .
RUN apk -q --no-cache add go build-base \
  && cd /root && go build -o /proxy \
  && cd / && rm -rf /root/* \
  && apk del --purge go build-base

ENTRYPOINT ["/proxy"]
