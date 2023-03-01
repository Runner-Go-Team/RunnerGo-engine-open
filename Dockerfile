FROM runnergo/debian:stable-slim

ADD  engine  /data/engine/engine

CMD ["/data/engine/engine","-m", "1"]
