.PHONY: dev check integration generate build

dev:
	bun run dev
check:
	bun run check
integration:
	go test -tags=integration -race -shuffle=on -count=1 ./...
generate:
	python3 scripts/contract.py
	bun run generate
build:
	go build ./cmd/...
	bun run build

lint:
	go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.15.0 run ./...

vuln:
	go run golang.org/x/vuln/cmd/govulncheck@v1.9.0 ./...
