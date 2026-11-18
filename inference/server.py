"""Local, bounded model inference only. Authorization, jobs and marking stay in Go."""

from __future__ import annotations

import argparse
import base64
import gc
import hmac
import io
import importlib.metadata
import tomllib
from pathlib import Path
import json
import logging
import os
import threading
import time
from http.server import BaseHTTPRequestHandler, ThreadingHTTPServer
from typing import Any

from bubbles import MODEL as BUBBLE_MODEL, WEIGHTS as BUBBLE_WEIGHTS, classify
import hashlib

import torch
from PIL import Image
from transformers import AutoModelForImageTextToText, AutoProcessor, LightOnOcrForConditionalGeneration, LightOnOcrProcessor

MODELS = {
    "lightonai/LightOnOCR-2-1B": "be7c41313c30e940fe5571de59e5fabc779e00c8",
    "Qwen/Qwen3.5-2B": "15852e8c16360a2fea060d615a32b45270f8a8fc",
}
MODELS[BUBBLE_MODEL] = hashlib.sha256(BUBBLE_WEIGHTS.read_bytes()).hexdigest()
Image.MAX_IMAGE_PIXELS = 8_000_000
LOCK = threading.Lock()
MARK_LOCK = threading.Lock()
ACTIVE: dict[str, Any] = {}
LOG = logging.getLogger("vision-runtime")


def load_model(name: str) -> tuple[Any, Any]:
    if ACTIVE.get("name") == name:
        return ACTIVE["model"], ACTIVE["processor"]
    ACTIVE.clear()
    gc.collect()
    torch.cuda.empty_cache()
    revision = MODELS[name]
    model_class = LightOnOcrForConditionalGeneration if name.startswith("lightonai/") else AutoModelForImageTextToText
    LOG.info("Loading allowlisted model %s at %s", name, revision)
    processor_class = LightOnOcrProcessor if name.startswith("lightonai/") else AutoProcessor
    processor = processor_class.from_pretrained(name, revision=revision, trust_remote_code=False)
    model = model_class.from_pretrained(
        name, revision=revision, dtype=torch.bfloat16,
        trust_remote_code=False, use_safetensors=True, attn_implementation="sdpa",
    ).to("cuda").eval()
    ACTIVE.update(name=name, model=model, processor=processor)
    return model, processor


def infer(payload: dict[str, Any]) -> dict[str, Any]:
    name = payload.get("model")
    if name not in MODELS:
        raise ValueError("Use an allowlisted, revision-pinned model")
    messages = payload.get("messages")
    if not isinstance(messages, list) or len(messages) != 1 or messages[0].get("role") != "user":
        raise ValueError("Send one user message with one inline PNG or JPEG")
    content = messages[0].get("content", [])
    if not isinstance(content, list) or len(content) > 2:
        raise ValueError("Send one image and at most one instruction")
    images = [part for part in content if part.get("type") == "image_url"]
    texts = [part.get("text", "") for part in content if part.get("type") == "text"]
    if len(images) != 1 or any(not isinstance(t, str) or len(t) > (262144 if name==BUBBLE_MODEL else 8000) for t in texts):
        raise ValueError("Invalid image or instruction count")
    uri = images[0].get("image_url", {}).get("url", "")
    if not uri.startswith(("data:image/png;base64,", "data:image/jpeg;base64,")):
        raise ValueError("Only inline PNG/JPEG data is accepted; remote URLs and files are forbidden")
    data = base64.b64decode(uri.split(",", 1)[1], validate=True)
    if len(data) > 8 << 20:
        raise ValueError("Image exceeds 8 MiB")
    source = Image.open(io.BytesIO(data))
    if source.width * source.height > 8_000_000 or source.format not in ("PNG", "JPEG"):
        raise ValueError("Image exceeds the pixel/format limit")
    image = source.convert("RGB")
    if name==BUBBLE_MODEL:
        if len(texts)!=1: raise ValueError("Supply bounded patch coordinates")
        config=json.loads(texts[0])
        started=time.monotonic()
        result=classify(image,config["centers"],config["radius"])
        return {"model":name,"model_revision":MODELS[name],"runtime_version":"bubble-classifier-v1","choices":[{"message":{"role":"assistant","content":json.dumps(result)},"finish_reason":"stop"}],"duration_ms":round((time.monotonic()-started)*1000)}
    image.thumbnail((1540, 1540) if name.startswith("lightonai/") else (1600, 1600))
    if not name.startswith("lightonai/"):
        scale=min(4.0,1600/image.width,1600/image.height)
        if scale>1: image=image.resize((round(image.width*scale),round(image.height*scale)),Image.Resampling.BICUBIC)
    max_tokens = payload.get("max_tokens", 4096)
    if not isinstance(max_tokens, int) or not 1 <= max_tokens <= 4096:
        raise ValueError("max_tokens must be 1–4096")
    if name.startswith("lightonai/") and texts:
        raise ValueError("LightOnOCR uses image-only transcription; instructions need the vision model")
    started = time.monotonic()
    model, processor = load_model(name)
    parts: list[dict[str, Any]] = [{"type": "image", "image": image}]
    parts.extend({"type": "text", "text": text} for text in texts)
    inputs = processor.apply_chat_template(
        [{"role": "user", "content": parts}], add_generation_prompt=True,
        tokenize=True, return_dict=True, return_tensors="pt", **({"enable_thinking": False} if not name.startswith("lightonai/") else {}),
    )
    inputs = {k: v.to(device="cuda", dtype=torch.bfloat16) if v.is_floating_point() else v.to("cuda") for k, v in inputs.items()}
    prompt_tokens = inputs["input_ids"].shape[-1]
    with torch.inference_mode():
        generated = model.generate(**inputs, max_new_tokens=max_tokens, do_sample=False, max_time=120)
    tokens = generated[0, prompt_tokens:]
    eos = model.generation_config.eos_token_id
    eos_ids = eos if isinstance(eos, list) else [eos]
    complete = len(tokens) > 0 and int(tokens[-1]) in eos_ids
    return {
        "model": name, "model_revision": MODELS[name], "runtime_version": "inline-vision-v2",
        "choices": [{"message": {"role": "assistant", "content": processor.decode(tokens, skip_special_tokens=True)}, "finish_reason": "stop" if complete else "length"}],
        "usage": {"prompt_tokens": prompt_tokens, "completion_tokens": len(tokens)},
        "duration_ms": round((time.monotonic() - started) * 1000),
    }


class Handler(BaseHTTPRequestHandler):
    def log_message(self, *_args: Any) -> None:
        pass

    def send_json(self, status: int, value: dict[str, Any]) -> None:
        data = json.dumps(value).encode()
        self.send_response(status)
        self.send_header("Content-Type", "application/json")
        self.send_header("Cache-Control", "no-store")
        self.send_header("Content-Length", str(len(data)))
        self.end_headers()
        self.wfile.write(data)

    def authorized(self) -> bool:
        return hmac.compare_digest(self.headers.get("Authorization", ""), "Bearer " + os.environ["VISION_API_KEY"])

    def do_GET(self) -> None:
        if not self.authorized():
            self.send_json(401, {"error": "Authentication required"})
            return
        if self.path != "/health":
            self.send_json(404, {"error": "Not found"})
            return
        self.send_json(200, {"ready": torch.cuda.is_available(), "models": MODELS, "active_model": ACTIVE.get("name"), "busy": LOCK.locked() or MARK_LOCK.locked()})

    def do_POST(self) -> None:
        if not self.authorized():
            self.send_json(401, {"error": "Authentication required"})
            return
        if self.path != "/v1/chat/completions":
            self.send_json(404, {"error": "Not found"})
            return
        lock=LOCK
        acquired=False
        try:
            length = int(self.headers.get("Content-Length", "0"))
            if not 1 <= length <= 12 << 20:
                raise ValueError("Request exceeds 12 MiB")
            self.connection.settimeout(15)
            payload = json.loads(self.rfile.read(length))
            if not isinstance(payload, dict):
                raise ValueError("JSON object required")
            lock=MARK_LOCK if payload.get("model")==BUBBLE_MODEL else LOCK
            acquired=lock.acquire(blocking=False)
            if not acquired:
                self.send_json(503,{"error":"Inference capacity busy; retry the saved job later"})
                return
            self.send_json(200, infer(payload))
        except (ValueError, KeyError, TypeError, Image.DecompressionBombError):
            self.send_json(400, {"error": "Invalid bounded image inference request"})
        except (BrokenPipeError, ConnectionResetError, TimeoutError):
            LOG.info("Inference client disconnected; output was not acknowledged")
        except Exception as error:
            LOG.error("Model inference failed: %s", type(error).__name__)
            self.send_json(503, {"error": "Model inference failed; retain the source and check runtime configuration"})
        finally:
            if acquired:lock.release()


def main() -> None:
    parser = argparse.ArgumentParser()
    parser.add_argument("--port", type=int, default=58010)
    parser.add_argument("--host", choices=["127.0.0.1", "0.0.0.0"], default="127.0.0.1")
    parser.add_argument("--warm", choices=list(MODELS))
    args = parser.parse_args()
    logging.basicConfig(level=logging.INFO)
    manifest = tomllib.loads(Path(__file__).with_name("pyproject.toml").read_text())
    for dependency in manifest["project"]["dependencies"]:
        package, expected = dependency.split("==", 1)
        if importlib.metadata.version(package) != expected:
            raise SystemExit("Inference dependency mismatch; run uv sync --project inference --frozen")
    if not os.environ.get("VISION_API_KEY") or len(os.environ["VISION_API_KEY"]) < 24:
        raise SystemExit("Set a private VISION_API_KEY of at least 24 characters")
    if not torch.cuda.is_available():
        raise SystemExit("A CUDA GPU is required by this configured runtime")
    torch.set_num_threads(4)
    if args.warm and args.warm != BUBBLE_MODEL:
        load_model(args.warm)
    LOG.info("Inference runtime listening on loopback port %d", args.port)
    ThreadingHTTPServer((args.host, args.port), Handler).serve_forever()


if __name__ == "__main__":
    main()
