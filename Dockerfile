FROM golang:1.18 as builder
WORKDIR /app
COPY . .
RUN go build -o gotorrent .

FROM ubuntu:latest
WORKDIR /app
COPY --from=builder /app/gotorrent .

ENTRYPOINT [ "/app/gotorrent" ]
VOLUME [ "/app/config", "/app/cache", "/app/downloads" ]
