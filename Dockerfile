# syntax=docker/dockerfile:1

##
## Build
##
FROM golang:1.18-alpine AS build

WORKDIR /app

COPY go.mod ./
COPY go.sum ./
RUN go mod download

COPY . .

RUN go build -o /extension-datadog

##
## Runtime
##
FROM alpine:3.16

WORKDIR /

COPY --from=build /extension-datadog /extension-datadog

EXPOSE 8088

ENTRYPOINT ["/extension-datadog"]
