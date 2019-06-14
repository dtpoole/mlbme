FROM golang as builder
ENV GO111MODULE=on
WORKDIR /go/src/github.com/dtpoole/mlbme
COPY *.go go.mod go.sum  ./
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o mlbme .

WORKDIR /go/src/go-mlbam-proxy
RUN git clone https://github.com/jwallet/go-mlbam-proxy.git ./
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o go-mlbam-proxy .


FROM jfloff/alpine-python:3.7-slim
ENV USER=xxx
ENV PATH="/usr/local/bin:${PATH}"

RUN addgroup -g 1000 $USER && adduser -D -u 1000 -G $USER $USER
RUN echo "http://dl-cdn.alpinelinux.org/alpine/edge/community" >> /etc/apk/repositories; \
    echo "http://dl-cdn.alpinelinux.org/alpine/edge/main" >> /etc/apk/repositories ; \
    apk upgrade --no-cache musl ;
RUN chmod +x /entrypoint.sh; /entrypoint.sh \
  -p streamlink \
  -a bash \
  -a vlc \
  && echo
RUN find /usr/local/lib/pyenv/versions/*/ -depth \( -name '*.pyo' -o -name '*.pyc' -o -name 'test' -o -name 'tests' \) -exec rm -rf '{}' + ;
RUN ln -s /usr/local/lib/pyenv/versions/*/bin/streamlink /usr/local/bin

WORKDIR /app

COPY --from=builder /go/src/go-mlbam-proxy/go-mlbam-proxy /usr/local/bin
COPY --from=builder /go/src/github.com/dtpoole/mlbme/mlbme .
COPY config.json  ./

USER $USER

ENTRYPOINT [ "./mlbme" ]