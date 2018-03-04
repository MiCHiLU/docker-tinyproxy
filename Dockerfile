FROM golang:alpine as build
RUN apk --update add \
  git \
  make \
  && go get github.com/adnanh/webhook
RUN go get github.com/golang/dep/cmd/dep
WORKDIR /go/src/github.com/MiCHiLU/docker-tinyproxy
ADD . .
RUN go get ./...
RUN make run

FROM alpine:latest
RUN apk --update add \
  bash \
  tinyproxy \
  ;
COPY --from=build /go/bin/webhook /usr/local/bin/
COPY --from=build /go/src/github.com/MiCHiLU/docker-tinyproxy/run /opt/docker-tinyproxy/
ADD hooks.json /opt/webhook/

EXPOSE 8888
EXPOSE 9000
ENTRYPOINT ["webhook", "-verbose", "-hooks", "/opt/webhook/hooks.json", "-urlprefix"]
