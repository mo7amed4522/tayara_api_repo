FROM golang:alpine as builder

WORKDIR /go/src/github.com/fadi/tayara_api
COPY . .

RUN go env -w GO111MODULE=on \
    && go env -w GOPROXY=https://goproxy.cn,direct \
    && go env -w CGO_ENABLED=0 \
    && go env \
    && go mod tidy \
    && go build -o server .

FROM alpine:latest

LABEL MAINTAINER="fadi@fadi.com"
# Set time zone
ENV TZ=Asia/Shanghai
RUN apk update && apk add --no-cache tzdata openntpd \
    && ln -sf /usr/share/zoneinfo/$TZ /etc/localtime && echo $TZ > /etc/timezone

WORKDIR /go/src/github.com/fadi/tayara_api

COPY --from=0 /go/src/github.com/fadi/tayara_api/server ./
COPY --from=0 /go/src/github.com/fadi/tayara_api/resource ./resource/
COPY --from=0 /go/src/github.com/fadi/tayara_api/config.docker.yaml ./

# Mount directory: If using an SQLite database, example container command: `docker run -d -v /host_path/gva.db:/go/src/github.com/fadi/tayara_api/gva.db -p 8888:8888 --name gva-server-v1 gva-server:1.0`
# VOLUME ["/go/src/github.com/fadi/tayara_api"]

EXPOSE 8888
ENTRYPOINT ./server -c config.docker.yaml