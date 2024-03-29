FROM golang:1.16 as builder
ARG VERSION=dev
ENV GO111MODULE=on
WORKDIR /go/src/github.com/dtpoole/mlbme
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags "-X main.version=${VERSION}" -a -installsuffix cgo -o mlbme .

WORKDIR /go/src/go-mlbam-proxy
RUN git clone https://github.com/jwallet/go-mlbam-proxy.git ./
RUN go mod init
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o go-mlbam-proxy .


FROM jfloff/alpine-python:3.8-slim
ENV USER=xxx
ENV PATH="/app:${PATH}"

RUN set -ex; \
  addgroup -g 1000 $USER && adduser -D -u 1000 -G $USER $USER; \
  echo "http://dl-cdn.alpinelinux.org/alpine/edge/community" >> /etc/apk/repositories; \
  echo "http://dl-cdn.alpinelinux.org/alpine/edge/main" >> /etc/apk/repositories; \
  /bin/bash entrypoint.sh \
    -p streamlink==2.1.1 \
    -p urllib3==1.25.11 \
    -a tzdata && \
  find /usr/local/lib/pyenv/versions/$PYTHON_VERSION/ -depth \( -name '*.pyo' -o -name '*.pyc' -o -name 'test' -o -name 'tests' -o -name 'SelfTest' \) -exec rm -rf '{}' + ; \
  rm -rf /root/.cache/pip;

WORKDIR /app

COPY --from=builder /go/src/go-mlbam-proxy/go-mlbam-proxy /usr/local/bin
COPY --from=builder /go/src/github.com/dtpoole/mlbme/mlbme .
COPY config.json  ./

USER $USER
EXPOSE 6789/tcp
ENTRYPOINT [ "mlbme" ]
