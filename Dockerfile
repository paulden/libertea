## Build
FROM golang:1.23-alpine AS build

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o libertea .

## Run
FROM alpine:edge

RUN apk --no-cache add ca-certificates tzdata

RUN addgroup -g 1000 sre && adduser -u 1000 -G sre -D sre

USER 1000

WORKDIR /app

COPY --chown=1000:1000 --from=build /app/libertea .

ENTRYPOINT ["/app/libertea"]
