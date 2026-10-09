## Build: cross-compile natively for the target platform, no emulation needed.
FROM --platform=$BUILDPLATFORM golang:1.27.2-alpine@sha256:85dc1069ac644ea3c527b177303a406eb3358192816cd7f9e5848eb658851673 AS build

ARG TARGETOS TARGETARCH
ARG VERSION=dev
ARG COMMIT=""
ARG DATE=""

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH \
    go build -trimpath \
      -ldflags="-s -w -X main.version=$VERSION -X main.commit=$COMMIT -X main.date=$DATE" \
      -o /out/libertea .

## Run: static binary on a distroless base, no shell nor package manager.
FROM gcr.io/distroless/static-debian13:nonroot@sha256:e2e927ec666bae08560abb3c55d0659eceabb657f56b6782ab500a9fc7f555e3

COPY --from=build /out/libertea /usr/local/bin/libertea

ENTRYPOINT ["/usr/local/bin/libertea"]
