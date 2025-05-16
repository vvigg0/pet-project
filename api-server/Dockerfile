#syntax=docker/dockerfile:1
FROM golang:1.24-alpine AS build-stage
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o api-server .
#Tests
FROM build-stage AS run-test-stage
RUN go test -v ./...
#Release build
FROM alpine:3.19
WORKDIR /
COPY --from=build-stage /app/api-server /api-server
EXPOSE 8080
CMD ["/api-server"]
