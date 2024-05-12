FROM golang:1.22.1-alpine as builder
WORKDIR /storage
COPY ../../go.sum go.mod ./
RUN go mod download
COPY ../.. /storage
RUN go build -o app ./cmd/storage/main.go

FROM alpine
WORKDIR /storage
COPY --from=builder /storage/app ./
COPY ../../migrations ./migrations

EXPOSE 8000
ENTRYPOINT ["/storage/app"]