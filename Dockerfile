FROM runnergo/debian:12-slim

ADD  engine  /data/engine/engine

CMD ["/data/engine/engine","-m", "1"]
