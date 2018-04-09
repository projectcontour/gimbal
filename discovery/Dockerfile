FROM alpine:3.6
LABEL maintainer="Steve Sloka <steves@heptio.com>"

RUN apk add --update ca-certificates && \
  rm -rf /var/cache/apk/*

ADD _output/bin/linux/amd64/kubernetes-discoverer /kubernetes-discoverer
ADD _output/bin/linux/amd64/openstack-discoverer /openstack-discoverer

USER nobody:nobody

ENTRYPOINT [ "/kubernetes-discoverer" ]