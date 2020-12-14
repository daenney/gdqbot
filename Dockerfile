FROM golang:1.15-buster as build

WORKDIR /go/src/app
ADD . /go/src/app

RUN go build -o /go/bin/gdqbot

FROM gcr.io/distroless/base-debian10:nonroot-amd64
COPY /go/bin/gdqbot /
ENTRYPOINT ["/gdqbot"]

LABEL \
  org.opencontainers.image.licenses="AGPL-3.0-or-later" \
  org.opencontainers.image.source="https://github.com/daenney/gdqbot" \
  org.opencontainers.image.title="gdqbot"
