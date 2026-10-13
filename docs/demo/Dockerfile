# Records the README demos in a real kitty terminal, so that the stratagem
# icons are shown: kitty runs in a virtual X server, xdotool types the keys
# and ffmpeg records the screen. Run from the repository root:
#
#   docker build -f docs/demo/Dockerfile -t libertea-demo .
#   docker run --rm -v "$PWD/docs/demo:/out" libertea-demo

FROM golang:1.27.2-alpine@sha256:85dc1069ac644ea3c527b177303a406eb3358192816cd7f9e5848eb658851673 AS build

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -trimpath -o /out/libertea .

FROM debian:trixie-slim@sha256:a29215f6a35e51e22adffa17f89e9d2ef06214e64a2bad10d765c46aea49f11f

RUN apt-get update \
    && apt-get install -y --no-install-recommends \
        ffmpeg \
        fonts-jetbrains-mono \
        kitty \
        libegl1 \
        libgl1-mesa-dri \
        xdotool \
        xvfb \
    && rm -rf /var/lib/apt/lists/*

# kitty needs OpenGL, rendered by Mesa on the CPU.
ENV LIBGL_ALWAYS_SOFTWARE=1 DISPLAY=:99 HOME=/tmp

COPY --from=build /out/libertea /usr/local/bin/libertea
COPY docs/demo/*.yaml docs/demo/record.sh /demo/

WORKDIR /demo
CMD ["/demo/record.sh", "/out"]
