# Go builder container
FROM golang:alpine as builder
# ENV GOLANG_VERSION 1.10.6

RUN apk add --no-cache make gcc musl-dev linux-headers

ADD . /go/src/github.com/scdoproject/go-stem

WORKDIR /go/src/github.com/scdoproject/go-stem

RUN make all

# Alpine container
FROM alpine:latest

RUN apk add --no-cache ca-certificates

COPY --from=builder /go/src/github.com/scdoproject/go-stem/build /go-stem

ENV PATH /go-stem:$PATH

RUN chmod +x /go-stem/node

EXPOSE 8027 8037 8057

# start a node with your 'config.json' file, this file must be external from a volume
# For example:
#   docker run -v <your config path>:/go-stem/config:ro -it go-stem node start -c /go-stem/config/configfile
