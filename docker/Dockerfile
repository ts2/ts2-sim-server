FROM debian:jessie
MAINTAINER Nicolas Piganeau <npi@m4x.org>

RUN mkdir /ts2

COPY ts2-sim-server /ts2/

COPY demo.json /ts2/

CMD ["/ts2/ts2-sim-server", "-loglevel", "debug", "/ts2/demo.json"]
