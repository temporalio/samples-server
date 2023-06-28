############################# Main targets #############################
all: clean build
########################################################################


##### Variables ######

COLOR := "\e[1;36m%s\e[0m\n"

ifndef GOOS
GOOS := $(shell go env GOOS)
endif

ifndef GOARCH
GOARCH := $(shell go env GOARCH)
endif

##### Build #####

build:
	@printf $(COLOR) "Building Kibble with OS: $(GOOS), ARCH: $(GOARCH)..."
	CGO_ENABLED=0 go build -ldflags "-s -w" ./cmd/kibble

clean:
	@printf $(COLOR) "Clearing binaries..."
	@rm -f kibble

##### Test #####
test:
	@printf $(COLOR) "Running unit tests..."
	go test ./... -race -count 1

##### Misc #####

update-dependencies:
	@printf $(COLOR) "Update dependencies..."
	@go get -u -t $(PINNED_DEPENDENCIES) ./...
	@go mod tidy

lint:
	@printf $(COLOR) "Run linters..."
	@golangci-lint run --verbose --timeout 10m --fix=false --new-from-rev=HEAD~ --config=.golangci.yml
