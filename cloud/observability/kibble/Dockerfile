FROM --platform=${BUILDPLATFORM:-linux/amd64} golang:1.20 as builder

ARG TARGETPLATFORM
ARG BUILDPLATFORM
ARG TARGETOS
ARG TARGETARCH

ARG Version
ARG GitCommit

WORKDIR ${GOPATH:-/go}/src/kibble

COPY . .
RUN go mod download
RUN go get -d -v ./...

RUN CGO_ENABLED=${CGO_ENABLED} GOOS=${TARGETOS} GOARCH=${TARGETARCH} \
  go build -o ${GOPATH:-/go}/bin/ ${GOPATH:-/go}/src/kibble/cmd/kibble

FROM --platform=${BUILDPLATFORM:-linux/amd64} centos:latest

COPY --from=builder ${GOPATH:-/go}/bin/kibble /

CMD ["/kibble"]
