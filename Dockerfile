FROM golang:1.15.2-alpine3.12 as build
RUN apk add git

WORKDIR /go/src/github.com/wiley/healthcat/

COPY go.mod .
COPY go.sum .

RUN go mod download -x

# If MODULE_NAME is not provided by the build process 
# version settings will not be set.
ARG MODULE_NAME

#disable crosscompiling
ENV CGO_ENABLED ${CGO_ENABLED:-0}
#Linux only
ENV GOOS ${GOOS:-linux}

COPY . .

RUN REPO_VERSION=$(git describe --abbrev=0 --tags) \
    BUILD_DATE=$(date +%Y-%m-%d-%H:%M) \
    GIT_HASH=$(git rev-parse --short HEAD) \
    go test ./...  \
    && go build -o healthcat \
        -ldflags "-X ${MODULE_NAME}/version.Version=${REPO_VERSION} \
                    -X ${MODULE_NAME}/version.GitCommit=${GIT_HASH} \
                    -X ${MODULE_NAME}/version.BuildDate=${BUILD_DATE}"

FROM scratch
COPY --from=build /go/src/github.com/wiley/healthcat/healthcat ./healthcat
COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
ENTRYPOINT ["./healthcat"]
