.PHONY: build clean

GO=CGO_ENABLED=1 GO111MODULE=on go

build:
	$(GO) build -o app-service ./cmd/edgex-nats-transport/main.go

docker:
	docker build \
	--build-arg http_proxy \
	--build-arg https_proxy \
	-f Dockerfile \
	-t edgexfoundry/docker-aws-export-example:dev \
	.

clean:
	rm -f app-service
