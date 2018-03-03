FROM golang:alpine as webhook
RUN apk --update add git \
  && go get github.com/adnanh/webhook

FROM alpine:latest
RUN apk --update add \
  bash \
  tinyproxy \
  ;
COPY --from=webhook /go/bin/webhook /usr/local/bin/
ADD run.sh /opt/docker-tinyproxy/run.sh
ADD hooks.json /opt/webhook/

EXPOSE 8888
EXPOSE 9000
ENTRYPOINT ["webhook", "-verbose", "-hooks", "/opt/webhook/hooks.json", "-urlprefix"]
