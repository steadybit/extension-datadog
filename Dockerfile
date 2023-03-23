# syntax=docker/dockerfile:1

##
## Build
##
FROM golang:1.19-alpine AS build

ARG NAME
ARG VERSION
ARG REVISION

WORKDIR /app

RUN apk add build-base
COPY go.mod ./
COPY go.sum ./
RUN go mod download

COPY . .

RUN go build \
        -ldflags="\
        -X 'github.com/steadybit/extension-kit/extbuild.ExtensionName=${NAME}' \
        -X 'github.com/steadybit/extension-kit/extbuild.Version=${VERSION}' \
        -X 'github.com/steadybit/extension-kit/extbuild.Revision=${REVISION}'" \
        -o /extension-datadog

##
## Runtime
##
FROM alpine:3.16

ARG USERNAME=steadybit
ARG USER_UID=1000

RUN adduser -u $USER_UID -D $USERNAME

USER $USERNAME

WORKDIR /

COPY --from=build /extension-datadog /extension-datadog

EXPOSE 8090

ENTRYPOINT ["/extension-datadog"]
