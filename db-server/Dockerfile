#syntax=docker/dockerfile:1
FROM golang:1.24-alpine AS build-stage
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o db-server .
#Tests
FROM build-stage AS run-test-stage
RUN go test -v ./...
#Release build
FROM alpine:3.19
WORKDIR /
COPY --from=build-stage /app/db-server /db-server
EXPOSE 8090
ENTRYPOINT ["/db-server"]
