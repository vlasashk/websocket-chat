FROM golang:1.22.1-alpine as builder
WORKDIR /server
COPY ../../go.sum go.mod ./
RUN go mod download
COPY ../.. /server
RUN go build -o app ./cmd/server/main.go

FROM alpine
WORKDIR /server
COPY --from=builder /server/app ./

EXPOSE 8080
ENTRYPOINT ["/server/app"]