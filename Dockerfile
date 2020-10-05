FROM golang:1.15.2-alpine3.12 as build
COPY . /go/src/github.com/wiley/do-k8s-cluster-health-check/
WORKDIR /go/src/github.com/wiley/do-k8s-cluster-health-check/

# If MODULE_NAME is not provided by the build process 
# version settings will not be set.
ARG MODULE_NAME

#disable crosscompiling
ENV CGO_ENABLED ${CGO_ENABLED:-0}
#Linux only
ENV GOOS ${GOOS:-linux}

RUN apk add git

RUN REPO_VERSION=$(git describe --abbrev=0 --tags) \
    BUILD_DATE=$(date +%Y-%m-%d-%H:%M) \
    GIT_HASH=$(git rev-parse --short HEAD) \
    go test ./...  \
    && go build -o chc \
        -ldflags "-X ${MODULE_NAME}/version.Version=${REPO_VERSION} \
                    -X ${MODULE_NAME}/version.GitCommit=${GIT_HASH} \
                    -X ${MODULE_NAME}/version.BuildDate=${BUILD_DATE}"

FROM scratch
COPY --from=build /go/src/github.com/wiley/do-k8s-cluster-health-check/chc ./chc
COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
ENTRYPOINT ["./chc"]
