FROM --platform=${BUILDPLATFORM:-linux/amd64} golang:1.25 AS builder

ARG TARGETPLATFORM
ARG BUILDPLATFORM
ARG TARGETOS
ARG TARGETARCH

ARG Version
ARG GitCommit

WORKDIR /go/src/promql-to-scrape

COPY . .
RUN go mod download

RUN CGO_ENABLED=${CGO_ENABLED:-0} GOOS=${TARGETOS} GOARCH=${TARGETARCH} \
  go build -o /go/bin/ /go/src/promql-to-scrape/cmd/promql-to-scrape

FROM --platform=${BUILDPLATFORM:-linux/amd64} alpine:latest

COPY --from=builder /go/bin/promql-to-scrape /

ENTRYPOINT ["/promql-to-scrape"]
