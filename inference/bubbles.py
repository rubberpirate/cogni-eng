"""Original bubble-mark classifier. No question, learner or marking policy is used."""
from __future__ import annotations

import hashlib
import io
from pathlib import Path

import torch
from PIL import Image
from safetensors.torch import load_file
from torch import nn

MODEL = "learning/bubble-marks-v1"
WEIGHTS = Path(__file__).with_name("bubble-marks-v1.safetensors")


def network() -> nn.Module:
    return nn.Sequential(
        nn.Conv2d(1, 8, 3, padding=1), nn.ReLU(), nn.AvgPool2d(2),
        nn.Conv2d(8, 16, 3, padding=1), nn.ReLU(), nn.AvgPool2d(2),
        nn.Flatten(), nn.Linear(16 * 8 * 8, 32), nn.ReLU(), nn.Linear(32, 3),
    )


def classify(image: Image.Image, centers: list[dict], radius: float) -> dict:
    if not 1 <= len(centers) <= 4000 or not .001 <= radius <= .05:
        raise ValueError("Invalid bounded mark patches")
    model = network().eval()
    model.load_state_dict(load_file(str(WEIGHTS)))
    patches = []
    for point in centers:
        x, y = float(point["x"]), float(point["y"])
        if not 0 <= x <= 1 or not 0 <= y <= 1:
            raise ValueError("Mark center is outside the image")
        r = max(2, round(radius * image.width))
        cx, cy = round(x * image.width), round(y * image.height)
        if cx-r < 0 or cy-r < 0 or cx+r >= image.width or cy+r >= image.height:
            raise ValueError("Mark patch crosses the image edge")
        patch = image.crop((cx-r, cy-r, cx+r+1, cy+r+1)).convert("L").resize((32, 32), Image.Resampling.BICUBIC)
        patches.append(torch.tensor(list(patch.get_flattened_data()), dtype=torch.float32).reshape(1, 32, 32) / 255)
    with torch.inference_mode():
        logits = model(torch.stack(patches))
        top = logits.topk(2, dim=1)
    return {
        "classes": top.indices[:, 0].tolist(),
        "margins": (top.values[:, 0] - top.values[:, 1]).tolist(),
        "revision": hashlib.sha256(WEIGHTS.read_bytes()).hexdigest(),
    }
