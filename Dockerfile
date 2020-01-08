FROM golang:1.12.7 as build
LABEL maintainer="Steve Sloka <slokas@vmware.com>"

WORKDIR /gimbal

ENV GOPROXY=https://gocenter.io
COPY go.mod go.sum /gimbal/
RUN go mod download

COPY cmd cmd
COPY pkg pkg

RUN CGO_ENABLED=0 GOOS=linux GOFLAGS=-ldflags=-w go build -o /go/bin/kubernetes-discoverer -ldflags=-s -v github.com/projectcontour/gimbal/cmd/kubernetes-discoverer
RUN CGO_ENABLED=0 GOOS=linux GOFLAGS=-ldflags=-w go build -o /go/bin/openstack-discoverer -ldflags=-s -v github.com/projectcontour/gimbal/cmd/openstack-discoverer

FROM scratch AS final
COPY --from=build /go/bin/kubernetes-discoverer /go/bin/kubernetes-discoverer
COPY --from=build /go/bin/openstack-discoverer /go/bin/openstack-discoverer

USER nobody:nobody

ENTRYPOINT [ "/kubernetes-discoverer" ]