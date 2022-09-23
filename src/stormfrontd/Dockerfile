FROM golang:1.18.1-alpine

RUN apk update && apk add git

WORKDIR /stormfrontd-build
COPY src/stormfrontd /stormfrontd-build
RUN env GOOS=linux CGO_ENABLED=0 go build -v -o stormfrontd

FROM alpine:latest

RUN adduser --disabled-password stormfrontd

WORKDIR /home/stormfrontd

COPY --from=0 /stormfrontd-build/stormfrontd ./

RUN apk update \
    && apk add bash

SHELL ["/bin/bash", "-c"]

RUN mkdir /home/stormfrontd/data
COPY data/stormfrontd /home/stormfrontd/data

ADD src/stormfrontd/start-stormfrontd.sh /home/stormfrontd/start-stormfrontd.sh
RUN chmod u+x /home/stormfrontd/start-stormfrontd.sh

RUN mkdir -p /home/stormfrontd/data
ADD data/stormfrontd/config.json /home/stormfrontd/data/config.json

RUN chown -R stormfrontd:stormfrontd /home/stormfrontd

USER stormfrontd

WORKDIR /home/stormfrontd

CMD ["./start-stormfrontd.sh"]
