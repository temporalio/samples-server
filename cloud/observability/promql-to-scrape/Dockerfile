FROM --platform=${BUILDPLATFORM:-linux/amd64} golang:1.21 as builder

ARG TARGETPLATFORM
ARG BUILDPLATFORM
ARG TARGETOS
ARG TARGETARCH

ARG Version
ARG GitCommit

WORKDIR ${GOPATH:-/go}/src/promql-to-scrape

COPY . .
RUN go mod download
RUN go get -d -v ./...

RUN CGO_ENABLED=${CGO_ENABLED:-0} GOOS=${TARGETOS} GOARCH=${TARGETARCH} \
  go build -o ${GOPATH:-/go}/bin/ ${GOPATH:-/go}/src/promql-to-scrape/cmd/promql-to-scrape

FROM --platform=${BUILDPLATFORM:-linux/amd64} alpine:latest

COPY --from=builder ${GOPATH:-/go}/bin/promql-to-scrape /

ENTRYPOINT ["/promql-to-scrape"]
