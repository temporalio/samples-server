FROM golang:1.21-alpine

WORKDIR /usr/src/app

COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY . .
RUN go build -v -o /usr/local/bin/promql-to-scrape ./cmd/promql-to-scrape/main.go

ENTRYPOINT ["/usr/local/bin/promql-to-scrape"]