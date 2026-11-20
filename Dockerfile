FROM golang:1.27.4-bookworm AS compile
WORKDIR /source
COPY go.mod go.sum ./
RUN go mod download
COPY cmd ./cmd
COPY internal ./internal
RUN CGO_ENABLED=0 go build -trimpath -o /out/gateway ./cmd/gateway && CGO_ENABLED=0 go build -trimpath -o /out/worker ./cmd/worker && CGO_ENABLED=0 go build -trimpath -o /out/operator ./cmd/operator

FROM python:3.13-slim-bookworm AS runtime
COPY --from=ghcr.io/astral-sh/uv:0.13.9 /uv /usr/local/bin/uv
RUN apt-get update && apt-get install -y --no-install-recommends ca-certificates tzdata poppler-utils tesseract-ocr tesseract-ocr-hin && rm -rf /var/lib/apt/lists/*
WORKDIR /app
COPY omr/pyproject.toml omr/uv.lock ./omr/
RUN UV_PYTHON_DOWNLOADS=never uv sync --project omr --frozen --no-dev --python /usr/local/bin/python3
COPY omr/bridge.py ./omr/bridge.py
COPY omr/engine ./omr/engine
COPY --from=compile /out/ /app/
COPY migrations ./migrations
USER 65532:65532
EXPOSE 8080
ENTRYPOINT ["/app/gateway"]
