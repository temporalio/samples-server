FROM --platform=${BUILDPLATFORM:-linux/amd64} golang:1.20 as builder

ARG TARGETPLATFORM
ARG BUILDPLATFORM
ARG TARGETOS
ARG TARGETARCH

ARG Version
ARG GitCommit

WORKDIR ${GOPATH:-/go}/src/promql-to-dd-go

COPY . .
RUN go mod download
RUN go get -d -v ./...

RUN CGO_ENABLED=${CGO_ENABLED:-0} GOOS=${TARGETOS} GOARCH=${TARGETARCH} \
  go build -o ${GOPATH:-/go}/bin/ ${GOPATH:-/go}/src/promql-to-dd-go/cmd/promqltodd

FROM --platform=${BUILDPLATFORM:-linux/amd64} alpine:latest

COPY --from=builder ${GOPATH:-/go}/bin/promqltodd /

ENTRYPOINT ["/promqltodd"]
