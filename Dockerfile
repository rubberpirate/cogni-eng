FROM golang:1.27.4-bookworm AS compile
WORKDIR /source
COPY go.mod go.sum ./
RUN go mod download
COPY cmd ./cmd
COPY internal ./internal
RUN CGO_ENABLED=0 go build -trimpath -o /out/gateway ./cmd/gateway && CGO_ENABLED=0 go build -trimpath -o /out/worker ./cmd/worker && CGO_ENABLED=0 go build -trimpath -o /out/operator ./cmd/operator

FROM debian:bookworm-slim AS runtime
RUN apt-get update && apt-get install -y --no-install-recommends ca-certificates tzdata poppler-utils tesseract-ocr tesseract-ocr-hin && rm -rf /var/lib/apt/lists/*
WORKDIR /app
COPY --from=compile /out/ /app/
COPY migrations ./migrations
USER 65532:65532
EXPOSE 8080
ENTRYPOINT ["/app/gateway"]
