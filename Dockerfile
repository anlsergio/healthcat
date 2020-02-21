FROM golang:1.13.8-alpine3.11 as build
COPY . /go/src/github.com/wiley/do-k8s-cluster-health-check/
WORKDIR /go/src/github.com/wiley/do-k8s-cluster-health-check/
#disable crosscompiling
ENV CGO_ENABLED ${CGO_ENABLED:-0}
#Linux only
ENV GOOS ${GOOS:-linux}

RUN go test ./... && go build -o chc

FROM scratch
COPY --from=build /go/src/github.com/wiley/do-k8s-cluster-health-check/chc ./chc
ENTRYPOINT ["./chc"]
