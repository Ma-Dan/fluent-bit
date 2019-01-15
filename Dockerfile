FROM gcr.io/google-containers/debian-base-amd64:0.3.1 as builderc

# Fluent Bit version
ENV FLB_MAJOR 0
ENV FLB_MINOR 14
ENV FLB_PATCH 9
ENV FLB_VERSION 0.14.9

ENV DEBIAN_FRONTEND noninteractive

RUN mkdir -p /fluent-bit/bin /fluent-bit/etc /fluent-bit/log /tmp/src/

COPY . /tmp/src/

RUN rm -rf /tmp/src/build/*

RUN apt-get update && \
    apt-get dist-upgrade -y && \
    apt-get install -y \
      build-essential \
      cmake \
      make \
      wget \
      unzip \
      libsystemd-dev \
      libssl1.0-dev \
      libasl-dev \
      libsasl2-dev

WORKDIR /tmp/src/build/
RUN cmake -DFLB_DEBUG=On \
          -DFLB_TRACE=Off \
          -DFLB_JEMALLOC=On \
          -DFLB_BUFFERING=On \
          -DFLB_TLS=On \
          -DFLB_WITHOUT_SHARED_LIB=On \
          -DFLB_WITHOUT_EXAMPLES=On \
          -DFLB_HTTP_SERVER=On \
          -DFLB_OUT_KAFKA=On ..
RUN make
RUN install bin/fluent-bit /fluent-bit/bin/

# Configuration files
COPY conf/fluent-bit.conf \
     conf/fluent-bit-custom.conf \
     conf/parsers.conf \
     conf/parsers_java.conf \
     conf/parsers_extra.conf \
     conf/parsers_openstack.conf \
     conf/parsers_cinder.conf \
     /fluent-bit/etc/


FROM golang:1.10.1-alpine3.7 as buildergo
WORKDIR /go/src
COPY fluentbitdaemon.go .

RUN apk update && apk add git
RUN go get github.com/golang/glog
RUN CGO_ENABLED=0 go build -o fluentbitdaemon ./fluentbitdaemon.go


FROM gcr.io/google-containers/debian-base-amd64:0.3.1
MAINTAINER Eduardo Silva <eduardo@treasure-data.com>
LABEL Description="Fluent Bit docker image" Vendor="Fluent Organization" Version="1.1"

RUN apt-get update \
    && apt-get dist-upgrade -y \
    && apt-get install --no-install-recommends ca-certificates libssl1.0.2 libsasl2-2 -y \
    && rm -rf /var/lib/apt/lists/* \
    && apt-get autoclean
COPY --from=builderc /fluent-bit /fluent-bit
COPY --from=buildergo /go/src/fluentbitdaemon /fluent-bit/bin/fluentbitdaemon

#
EXPOSE 2020
EXPOSE 24444

# Entry point
CMD ["/fluent-bit/bin/fluentbitdaemon"]
