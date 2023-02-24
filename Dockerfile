FROM runnergo/debian:stable-slim

ADD  engine  /data/engine/engine


ADD  wait-for-it.sh /bin/

CMD ["/data/engine/engine","-m", "1"]
