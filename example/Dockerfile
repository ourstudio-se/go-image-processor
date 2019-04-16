FROM golang:1.12-alpine as build-env

RUN apk add --no-cache \
	ca-certificates \
	git \
    alpine-sdk \
    imagemagick-dev

WORKDIR /go/src/app
COPY . .

RUN go get -d
RUN go build -o $GOBIN/go-image-processor-api

FROM alpine

WORKDIR /app
COPY --from=build-env $GOBIN/go-image-processor-api /app/

RUN apk add --no-cache \
    ca-certificates \
    imagemagick \
    imagemagick-dev \
    ghostscript \
    libwebp-dev \
    libjpeg-turbo \
    jasper \
    libpng \
    tiff \
    libraw

ENTRYPOINT ./go-image-processor-api