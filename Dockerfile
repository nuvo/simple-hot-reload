FROM golang:1.12-alpine3.9 as builder

WORKDIR /workdir

RUN apk add --no-cache git

COPY hot-reload.go /workdir/hot-reload.go

RUN go get -d ./... && \
    go build hot-reload.go

FROM alpine:3.9

WORKDIR /app

COPY --from=builder /workdir/hot-reload /app/hot-reload

ENTRYPOINT ["./hot-reload"]
