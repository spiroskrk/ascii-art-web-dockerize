FROM golang:1.22.2-alpine AS builder

WORKDIR /src

COPY go.mod ./
RUN go mod download

COPY . .

RUN go test ./...
RUN mkdir -p /out && \
    CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/server .

FROM debian:bookworm-slim AS runtime

LABEL org.opencontainers.image.title="ASCII Art Web" \
      org.opencontainers.image.description="A containerized Go web application that renders ASCII art" \
      org.opencontainers.image.authors="Aris Kasapidis, Spiros, Kostis"

RUN groupadd --system app && \
    useradd --system --gid app --home-dir /app --no-create-home app

WORKDIR /app

COPY --from=builder --chown=app:app /out/server ./server
COPY --chown=app:app banners ./banners
COPY --chown=app:app templates ./templates
COPY --chown=app:app static ./static

USER app

EXPOSE 8080
STOPSIGNAL SIGTERM

ENTRYPOINT ["./server"]
