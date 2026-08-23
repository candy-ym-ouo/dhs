fmt:
	gofmt -w $$(find . -name '*.go')
vet:
	go vet ./...
build:
	mkdir -p bin && go build -o bin/heartbeat ./cmd/heartbeat
test:
	go test ./... -race -count=1
size:
	bash scripts/check_size.sh
verify: fmt vet build test size
