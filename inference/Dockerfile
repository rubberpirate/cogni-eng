FROM python:3.13.18-slim
COPY --from=ghcr.io/astral-sh/uv:0.13.6 /uv /bin/uv
WORKDIR /runtime
COPY pyproject.toml uv.lock ./
RUN uv sync --frozen --no-dev
COPY server.py bubbles.py bubble-marks-v1.safetensors ./
ENV HF_HOME=/models HF_HUB_DISABLE_TELEMETRY=1 DO_NOT_TRACK=1
EXPOSE 58010
CMD ["/runtime/.venv/bin/python", "server.py", "--host", "0.0.0.0"]
