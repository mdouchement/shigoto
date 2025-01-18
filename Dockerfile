# build stage
FROM golang:alpine AS build-env
LABEL maintainer="mdouchement"

RUN apk upgrade
RUN apk add --update --no-cache git curl

ARG CHECKSUM_VERSION=v0.2.4
ARG CHECKSUM_SUM=0540c8446174873f140a8fe8c14d7db46c02d81e1e936b0019fcfa121258e12b

RUN curl -L https://github.com/mdouchement/checksum/releases/download/$CHECKSUM_VERSION/checksum-linux-amd64 -o /usr/local/bin/checksum && \
    echo "$CHECKSUM_SUM  /usr/local/bin/checksum" | sha256sum -c && \
    chmod +x /usr/local/bin/checksum

ARG TASK_VERSION=v3.41.0
ARG TASK_SUM=0a2595f7fa3c15a62f8d0c244121a4977018b3bfdec7c1542ac2a8cf079978b8

RUN curl -LO https://github.com/go-task/task/releases/download/$TASK_VERSION/task_linux_amd64.tar.gz && \
    checksum --verify=$TASK_SUM task_linux_amd64.tar.gz && \
    tar -xf task_linux_amd64.tar.gz && \
    cp task /usr/local/bin/

RUN mkdir -p /go/src/github.com/mdouchement/shigoto
WORKDIR /go/src/github.com/mdouchement/shigoto

ENV CGO_ENABLED=0
ENV GOPROXY=https://proxy.golang.org

COPY . /go/src/github.com/mdouchement/shigoto

RUN go mod download
RUN task build-for-docker

# final stage
FROM alpine
LABEL maintainer="mdouchement"

COPY --from=build-env /go/src/github.com/mdouchement/shigoto/dist/shigoto /usr/local/bin/

RUN mkdir -p /var/run
RUN mkdir -p /etc/shigoto
RUN mkdir -p /etc/shigoto.d # YAML directory

COPY <<EOF /etc/shigoto/shigoto.toml
directory = "/etc/shigoto.d"
socket = "/var/run/shigoto.sock"

[log]
force_color = true
force_formating = true
EOF

CMD ["shigoto", "daemon"]
