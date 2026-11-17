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
