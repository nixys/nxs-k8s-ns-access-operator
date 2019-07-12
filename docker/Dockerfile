FROM golang:1.10 as build-env

WORKDIR /go/src/nxs-k8s-ns-access-operator
ADD ./src /go/src/nxs-k8s-ns-access-operator

RUN go get -d -v
RUN go install

FROM gcr.io/distroless/base
COPY --from=build-env /go/bin/nxs-k8s-ns-access-operator /
CMD ["/nxs-k8s-ns-access-operator"]
