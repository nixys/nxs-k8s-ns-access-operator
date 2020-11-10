FROM golang:1.14 as build-env

WORKDIR /usr/src/nxs-k8s-ns-access-operator
ADD ./ /usr/src/nxs-k8s-ns-access-operator

RUN go build -o /nxs-k8s-ns-access-operator

FROM gcr.io/distroless/base

COPY --from=build-env /nxs-k8s-ns-access-operator /

CMD ["/nxs-k8s-ns-access-operator"]
