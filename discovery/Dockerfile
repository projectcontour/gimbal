FROM debian:stretch-slim
LABEL maintainer="Steve Sloka <slokas@vmware.com>"

RUN apt-get update && apt-get install -y ca-certificates && \
  apt-get clean autoclean && apt-get autoremove -y && \
  rm -rf /var/lib/{apt,dpkg,cache,log}/

ADD _output/bin/linux/amd64/kubernetes-discoverer /kubernetes-discoverer
ADD _output/bin/linux/amd64/openstack-discoverer /openstack-discoverer

USER nobody:nobody

ENTRYPOINT [ "/kubernetes-discoverer" ]