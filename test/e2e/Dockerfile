# syntax=docker/dockerfile:1

FROM golang:1.23-bookworm

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 go build \
    -cover \
    -covermode=atomic \
    -coverpkg=./... \
    -trimpath \
    -ldflags="-X main.version=e2e-version -X main.architecture=e2e-architecture" \
    -o /usr/local/bin/mdefaults \
    ./cmd/mdefaults

RUN CGO_ENABLED=0 go test \
    -c \
    -tags=e2e \
    -o /usr/local/bin/mdefaults-e2e.test \
    ./test/e2e

RUN useradd --create-home --uid 10001 --shell /usr/sbin/nologin e2e

USER 10001:10001
WORKDIR /tmp

ENV GOCACHE=/tmp/go-build-cache \
    GOTELEMETRY=off \
    GOTOOLCHAIN=local \
    HOME=/tmp/e2e-runner-home \
    MDEFAULTS_E2E_BINARY=/usr/local/bin/mdefaults \
    MDEFAULTS_E2E_CONTAINER=1 \
    MDEFAULTS_E2E_COVERAGE_DIR=/tmp/mdefaults-e2e-coverage

ENTRYPOINT ["/usr/local/bin/mdefaults-e2e.test", "-test.v", "-test.count=1"]
