FROM golang:1.13.5 as build
LABEL maintainer="Steve Sloka <slokas@vmware.com>"

WORKDIR /gimbal

ENV GOPROXY=https://proxy.golang.org
COPY go.mod go.sum /gimbal/
RUN go mod download

COPY cmd cmd
COPY pkg pkg

RUN CGO_ENABLED=0 GOOS=linux GOFLAGS=-ldflags=-w go build -o /go/bin/kubernetes-discoverer -ldflags=-s -v github.com/projectcontour/gimbal/cmd/kubernetes-discoverer
RUN CGO_ENABLED=0 GOOS=linux GOFLAGS=-ldflags=-w go build -o /go/bin/openstack-discoverer -ldflags=-s -v github.com/projectcontour/gimbal/cmd/openstack-discoverer

FROM scratch AS final
COPY --from=build /go/bin/kubernetes-discoverer /kubernetes-discoverer
COPY --from=build /go/bin/openstack-discoverer /openstack-discoverer

ENTRYPOINT [ "/kubernetes-discoverer" ]