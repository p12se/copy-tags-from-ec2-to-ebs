#build stage

FROM golang:1.11.2-stretch AS builder
RUN  mkdir -p /go/src \
    && mkdir -p /go/bin \
    && mkdir -p /go/pkg

ENV GOPATH=/go
ENV PATH=$GOPATH/bin:$PATH
ENV GOOS=linux
ENV GOARCH=amd64

RUN mkdir -p $GOPATH/src/app 
ADD . $GOPATH/src/app

WORKDIR $GOPATH/src/app

RUN GOOS=linux GOARCH=amd64 go test -v .
RUN GOOS=linux GOARCH=amd64 go build -o main
RUN GOOS=linux GOARCH=amd64 go build ./vendor/github.com/aws/aws-lambda-go/cmd/build-lambda-zip

RUN $GOPATH/src/app/build-lambda-zip -o main.zip main

#final stage
FROM alpine:latest
ENV GOPATH=/go
ENV PATH=$GOPATH/bin:$PATH
RUN mkdir /app

COPY --from=builder $GOPATH/src/app/main.zip /app/main.zip

LABEL Name=copy-tags-from-ec2-to-ebs Version=0.0.1
