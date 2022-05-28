FROM golang:alpine

MAINTAINER Maintainer

ENV GIN_MODE=debug
ENV PORT=8000
ENV MODE=DOCKER

RUN apk add build-base
WORKDIR kafka-test

COPY . .

RUN go mod download
RUN go build -o app-exe ./cmd/...

EXPOSE $PORT

CMD ["./app-exe"]
