FROM golang:1.18.1-alpine

RUN apk update && apk add git

WORKDIR /stormfrontd-build
COPY src/stormfrontd /stormfrontd-build
RUN env GOOS=linux CGO_ENABLED=0 go build -v -o stormfrontd



FROM golang:1.18.1-alpine

RUN apk update && apk add git

WORKDIR /stormfront-cli-build
COPY src/stormfront-cli /stormfront-cli-build
RUN env GOOS=linux CGO_ENABLED=0 go build -v -o stormfront



FROM alpine:latest

RUN adduser --disabled-password stormfrontd

WORKDIR /home/stormfrontd

COPY --from=0 /stormfrontd-build/stormfrontd ./

COPY --from=1 /stormfront-cli-build/stormfront ./

RUN apk update \
    && apk add bash

SHELL ["/bin/bash", "-c"]

RUN chown -R stormfrontd:stormfrontd /home/stormfrontd

USER 0

WORKDIR /home/stormfrontd

CMD ["./stormfrontd"]
