"""Reproduce original synthetic mark training; never reads the owner's evaluation sheet."""
from __future__ import annotations

import io
import hashlib
import json
import random
import time
from pathlib import Path

import numpy as np
import torch
from PIL import Image, ImageDraw, ImageFilter, ImageFont
from safetensors.torch import save_file

from bubbles import WEIGHTS, network


def sample(rng: random.Random, label: int) -> np.ndarray:
    background = rng.randint(232, 255)
    image = Image.new("RGB", (64, 64), (background, background, max(220, background-rng.randint(0, 8))))
    draw = ImageDraw.Draw(image)
    cx, cy = 32+rng.randint(-3, 3), 32+rng.randint(-3, 3)
    radius = rng.randint(20, 27)
    ink = rng.choice([(rng.randint(150, 215), rng.randint(80, 155), rng.randint(80, 155)), (100, 100, 100), (50, 50, 50)])
    draw.ellipse((cx-radius, cy-radius, cx+radius, cy+radius), outline=ink, width=rng.randint(1, 3))
    font = ImageFont.load_default(size=rng.randint(21, 30))
    draw.text((cx, cy), str(rng.randint(1, 8)), fill=ink, font=font, anchor="mm")
    if label:
        fill = rng.choice([(rng.randint(0, 65), rng.randint(0, 65), rng.randint(0, 65)), (15, 20, 75)])
        r = radius*rng.uniform(.72, 1.0)
        box = (cx-r, cy-r, cx+r, cy+r)
        if label == 1:
            draw.ellipse(box, fill=fill)
        else:
            draw.pieslice(box, rng.randint(0, 90), rng.randint(140, 290), fill=fill)
    image = image.rotate(rng.uniform(-8, 8), fillcolor=(background,)*3)
    image = image.filter(ImageFilter.GaussianBlur(rng.uniform(0, 1.1)))
    if rng.random() < .5:
        encoded=io.BytesIO();image.save(encoded,format="JPEG",quality=rng.randint(45,95));image=Image.open(io.BytesIO(encoded.getvalue())).convert("RGB")
    return np.array(image.convert("L").resize((32,32),Image.Resampling.BICUBIC),dtype=np.float32)[None]/255


def main() -> None:
    torch.manual_seed(42017)
    rng=random.Random(42017)
    torch.set_num_threads(4)
    started=time.monotonic()
    labels=torch.tensor([i%3 for i in range(12000)],dtype=torch.long)
    images=torch.tensor(np.stack([sample(rng,int(label)) for label in labels]))
    device="cuda" if torch.cuda.is_available() else "cpu"
    model=network().to(device);optimizer=torch.optim.Adam(model.parameters(),lr=.001)
    train_x,train_y=images[:10000].to(device),labels[:10000].to(device)
    for epoch in range(16):
        order=torch.randperm(len(train_x),device=device)
        model.train()
        for start in range(0,len(order),256):
            ix=order[start:start+256];optimizer.zero_grad();loss=nn_loss(model(train_x[ix]),train_y[ix]);loss.backward();optimizer.step()
        print("epoch",epoch+1,"loss",round(float(loss.detach()),5),flush=True)
    model.eval()
    with torch.inference_mode():
        prediction=model(images[10000:].to(device)).argmax(1).cpu()
    correct=int((prediction==labels[10000:]).sum())
    state={k:v.cpu().contiguous() for k,v in model.state_dict().items()};save_file(state,str(WEIGHTS))
    report={"weight_sha256":hashlib.sha256(WEIGHTS.read_bytes()).hexdigest(),"dataset":"Original synthetic circle/digit patches; no owner images used for training", "seed":42017,"training_patches":10000,"synthetic_holdout_patches":2000,"synthetic_holdout_correct":correct,"classes":{"0":"unfilled printed bubble","1":"filled bubble","2":"partial/ambiguous fill"},"epochs":16,"device":device,"seconds":round(time.monotonic()-started,2),"limitations":"Synthetic validation only. Not a field accuracy claim or calibrated probability. Real photos, erasures, diverse pens/layouts and calibration errors require broader evaluation."}
    WEIGHTS.with_suffix(".training.json").write_text(json.dumps(report,indent=2)+"\n")
    print(json.dumps(report),flush=True)


def nn_loss(logits: torch.Tensor, labels: torch.Tensor) -> torch.Tensor:
    return torch.nn.functional.cross_entropy(logits,labels)


if __name__=="__main__":
    main()
