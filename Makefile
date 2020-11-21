OUTPUT=build/bin

.PHONY: lint unittest build clean

lint:
	@golangci-lint run  -v cmd/... internal/...

unittest:
	@go test -v ./internal/... 

build: lint unittest
	@go build -o $(OUTPUT)/api_server cmd/main.go

clean: 
	@ rm -rf build/

start_redis:
	docker run -p 6379:6379  --name demo_redis -d redis:6.0

stop_redis:
	docker rm -f demo_redis
