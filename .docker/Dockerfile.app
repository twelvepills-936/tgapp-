FROM golang:1.24-bookworm AS build

ENV GO111MODULE=on

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN env CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -v -o /go/bin/service ./cmd/service

FROM alpine:3.20

ARG release
ENV APP_RELEASE=$release
ENV GOLANG_PROTOBUF_REGISTRATION_CONFLICT=warn

RUN apk add --no-cache bash ca-certificates

COPY --from=build /app/api/ /api

COPY --from=build /usr/local/go/lib/time/zoneinfo.zip /
ENV ZONEINFO=/zoneinfo.zip

COPY --from=build /go/bin/service /go/bin/service
