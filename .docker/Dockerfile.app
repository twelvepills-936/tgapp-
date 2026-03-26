FROM registry-gitlab16.skiftrade.kz/lib/service-golang:v2.1 as build

ENV GO111MODULE=on

WORKDIR ${GOPATH}/src/gitlab16.skiftrade.kz/templates/go

ENV DIR_PROJECT ${GOPATH}/src/gitlab16.skiftrade.kz/templates/go

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN env CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -v -o /go/bin/service ./cmd/service

FROM registry-gitlab16.skiftrade.kz/lib/images/alpine/alpine:3.20

ARG release
ENV APP_RELEASE=$release
ENV GOLANG_PROTOBUF_REGISTRATION_CONFLICT=warn

RUN apk add --no-cache bash

COPY --from=build /go/src/gitlab16.skiftrade.kz/templates/go/api/ /api
COPY --from=build /go/src/gitlab16.skiftrade.kz/templates/go/scripts/ /scripts

COPY --from=build /usr/local/go/lib/time/zoneinfo.zip /
ENV ZONEINFO=/zoneinfo.zip

COPY --from=build /go/bin/service /go/bin/service