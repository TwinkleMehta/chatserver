FROM golang:alpine as build-env

ENV GO111MODULE=on

RUN apk update && apk add bash ca-certificates git gcc g++ libc-dev

WORKDIR /go/src/github.com/TwinkleMehta/chatserver
COPY . .

RUN go mod download 

RUN go build -o chatserver

CMD ./chatserver