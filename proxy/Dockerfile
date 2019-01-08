ARG THIS_ARCH_ALT=amd64
FROM multiarch/alpine:${THIS_ARCH_ALT}-edge

LABEL Lan Tian "lantian@lantian.pub"
ENV GOOS=linux GOARCH=${THIS_ARCH_GO}
WORKDIR /root
COPY . .
RUN apk -q --no-cache add git go build-base upx \
  && cd /root && go build -o /proxy && upx /proxy \
  && cd / && rm -rf /root/* \
  && apk del go build-base upx

ENTRYPOINT ["/proxy"]
