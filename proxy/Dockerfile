ARG THIS_ARCH_ALT=amd64
FROM multiarch/alpine:${THIS_ARCH_ALT}-edge

LABEL Lan Tian "lantian@lantian.pub"
COPY build/proxy /
ENTRYPOINT ["/lib/ld-musl-x86_64.so.1 /proxy"]
