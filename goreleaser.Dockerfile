# Image built by GoReleaser from the binaries it already compiled, see .goreleaser.yaml.
FROM gcr.io/distroless/static-debian13:nonroot@sha256:e2e927ec666bae08560abb3c55d0659eceabb657f56b6782ab500a9fc7f555e3

ARG TARGETPLATFORM

COPY $TARGETPLATFORM/libertea /usr/local/bin/libertea

ENTRYPOINT ["/usr/local/bin/libertea"]
