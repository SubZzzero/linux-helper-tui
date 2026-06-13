build:
	go build -o bin/linux-helper ./cmd/linux-helper

test:
	go test ./... -race -count=1

lint:
	golangci-lint run

cover:
	go test ./... -coverprofile=coverage.out && go tool cover -html=coverage.out

bench:
	go test ./... -run '^$' -bench .

clean:
	rm -rf bin/ coverage.out
