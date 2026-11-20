"""Headless adapter to the bundled OMRChecker engine; no grading or roster access."""

from __future__ import annotations

import argparse
import base64
from contextlib import redirect_stderr, redirect_stdout
from copy import deepcopy
import json
import os
from pathlib import Path
import re
import shutil
import subprocess
import sys
import tempfile
from types import SimpleNamespace

sys.dont_write_bytecode = True


class InputError(ValueError):
    pass


def number_between(value: object, low: float, high: float, *, integer: bool = False) -> bool:
    return type(value) in ((int,) if integer else (int, float)) and low <= value <= high


def field_count(labels: list) -> int:
    if not isinstance(labels, list) or len(labels) > 1000:
        raise InputError("Use a bounded list of field labels.")
    count = 0
    for label in labels:
        if not isinstance(label, str) or not 1 <= len(label) <= 64:
            raise InputError("Field labels must contain 1–64 characters.")
        if ".." in label:
            match = re.fullmatch(r"[^\d.]+(\d+)\.\.(\d+)", label)
            if not match or not 1 <= int(match[2]) - int(match[1]) + 1 <= 500:
                raise InputError("Field ranges must expand to at most 500 fields.")
            count += int(match[2]) - int(match[1]) + 1
        else:
            count += 1
        if count > 1000:
            raise InputError("Field label lists may expand to at most 1000 entries.")
    return count


def bounded_template(template: dict, marker: bool, reference: bool) -> dict:
    if not isinstance(template, dict):
        raise InputError("Provide an OMRChecker template JSON object.")
    template = deepcopy(template)
    dimensions = template.get("pageDimensions", [])
    if not isinstance(dimensions,list) or len(dimensions) != 2 or any(type(n) is not int or not 100 <= n <= 5000 for n in dimensions) or dimensions[0] * dimensions[1] > 8_000_000:
        raise InputError("Template page dimensions must be 100–5000 pixels and at most 8 megapixels.")
    blocks = template.get("fieldBlocks")
    if not isinstance(blocks, dict) or not 1 <= len(blocks) <= 500:
        raise InputError("Configure 1–500 OMRChecker field blocks.")
    field_count(template.get("outputColumns", []))
    custom = template.get("customLabels", {})
    if not isinstance(custom, dict) or len(custom) > 500:
        raise InputError("Use at most 500 combined fields.")
    for labels in custom.values():
        field_count(labels)
    count = 0
    for block in blocks.values():
        if not isinstance(block, dict) or not isinstance(block.get("fieldLabels"), list):
            raise InputError("Each field block needs fieldLabels.")
        count += field_count(block["fieldLabels"])
        if count > 500:
            raise InputError("A template may contain at most 500 response fields.")
        values = block.get("bubbleValues", [])
        if not isinstance(values, list) or len(values) > 32 or any(not isinstance(v, str) or not 1 <= len(v) <= 16 for v in values):
            raise InputError("Use at most 32 short bubble values per field.")
    processors = template.get("preProcessors", [])
    if not isinstance(processors, list) or len(processors) > 8:
        raise InputError("Use at most eight preprocessing steps.")
    for processor in processors:
        if not isinstance(processor, dict):
            raise InputError("Invalid preprocessing step.")
        name = processor.get("name")
        if not isinstance(name,str):
            raise InputError("Each preprocessor needs a name.")
        options = processor.setdefault("options", {})
        if not isinstance(options, dict):
            raise InputError("Preprocessor options must be an object.")
        if name == "CropOnMarkers":
            if not marker:
                raise InputError("This template needs its marker image. Upload the marker with the sheet.")
            options["relativePath"] = "marker.png"
            steps = options.get("marker_rescale_steps", 10)
            scales = options.get("marker_rescale_range", [35,100])
            if not number_between(steps,1,100,integer=True) or not isinstance(scales,(list,tuple)) or len(scales)!=2 or any(type(n) is not int or not 1 <= n <= 200 for n in scales) or scales[0]>scales[1] or not number_between(options.get("sheetToMarkerWidthRatio",17),1,500):
                raise InputError("Marker alignment needs 1–100 scale steps within 1–200 percent and a positive sheet/marker ratio.")
        elif name == "FeatureBasedAlignment":
            if not reference:
                raise InputError("This template needs its blank reference image. Upload the reference with the sheet.")
            options["reference"] = "reference.png"
            if not number_between(options.get("maxFeatures", 500),1,5000,integer=True):
                raise InputError("Feature alignment must use at most 5000 features.")
        elif name not in {"CropPage", "Levels", "MedianBlur", "GaussianBlur"}:
            raise InputError("Unsupported preprocessor. Use CropPage, CropOnMarkers, FeatureBasedAlignment, Levels or blur filters.")
        for key in ("morphKernel", "kSize"):
            if key in options:
                kernel = options[key] if isinstance(options[key], list) else [options[key]]
                if not 1 <= len(kernel) <= 2 or any(type(n) is not int or not 1 <= n <= 51 for n in kernel):
                    raise InputError("Filter kernels must contain 1–51 pixels per dimension.")
        if name == "Levels" and not (number_between(options.get("low",0),0,1) and number_between(options.get("high",1),0,1) and options.get("low",0)<options.get("high",1) and number_between(options.get("gamma",1),.1,10)):
            raise InputError("Levels require low < high within 0–1 and gamma within 0.1–10.")
    return template


def decode_values(raw: str, options: list[str]) -> list[str] | None:
    """Only accept a unique ordered decomposition of upstream's concatenated marks."""
    possibilities: list[list[str]] = []

    def visit(index: int, rest: str, chosen: list[str]) -> None:
        if len(possibilities) > 1:
            return
        if not rest:
            possibilities.append(chosen)
            return
        for position in range(index, len(options)):
            value = options[position]
            if value and rest.startswith(value):
                visit(position + 1, rest[len(value):], [*chosen, value])

    visit(0, raw, [])
    return possibilities[0] if len(possibilities) == 1 else None


def read_identity(image, regions: list, directory: Path) -> list[dict]:
    import cv2

    if not isinstance(regions, list) or len(regions) > 2:
        raise InputError("Select at most one name and one roll-number field.")
    readings = []
    seen = set()
    height, width = image.shape[:2]
    for region in regions:
        if not isinstance(region, dict) or not isinstance(region.get("rect"), dict):
            raise InputError("Select identity fields as boxes inside the sheet.")
        kind, rect = region.get("kind"), region["rect"]
        if (
            kind not in {"student_name", "roll_number"}
            or kind in seen
            or any(not number_between(rect.get(k), 0, 1) for k in ("x", "y", "width", "height"))
            or rect["width"] <= 0 or rect["height"] <= 0
            or rect["x"] + rect["width"] > 1 + 1e-8
            or rect["y"] + rect["height"] > 1 + 1e-8
        ):
            raise InputError("Identity fields must be unique boxes inside the sheet.")
        seen.add(kind)
        crop = image[
            round(rect["y"] * height):round((rect["y"] + rect["height"]) * height),
            round(rect["x"] * width):round((rect["x"] + rect["width"]) * width),
        ]
        result = {"kind": kind, "text": "", "confidence": 0, "uncertain": True,
                  "reason": "No clear text; check this field or enter it manually."}
        if crop.size:
            path = directory / (kind + ".png")
            cv2.imwrite(str(path), crop)
            try:
                output = subprocess.run(
                    ["tesseract", str(path), "stdout", "-l", "eng", "--psm", "7", "tsv"],
                    capture_output=True, text=True, timeout=12, check=True,
                ).stdout
                words, confidences = [], []
                for line in output.splitlines()[1:]:
                    columns = line.split("\t", 11)
                    if len(columns) == 12 and columns[11].strip() and float(columns[10]) >= 0:
                        words.append(columns[11].strip())
                        confidences.append(float(columns[10]))
                text = re.sub(
                    r"^(?:student\s*name|name|roll(?:\s*(?:number|no\.?))?)\s*[:#]\s*", "",
                    " ".join(words), flags=re.I,
                )
                if kind == "roll_number":
                    text = text.replace(" ", "")
                confidence = sum(confidences) / len(confidences) if confidences else 0
                reason = "Check the detected text; OCR confidence is low." if text else result["reason"]
                result.update(text=text[:160], confidence=round(confidence, 2),
                              uncertain=not text or confidence < 85,
                              reason="" if text and confidence >= 85 else reason)
            except (OSError, subprocess.SubprocessError):
                result["reason"] = "Text recognition is unavailable. Enter this field manually."
        readings.append(result)
    return readings


def read_sheet(root: Path, image_path: Path, payload: dict) -> dict:
    # Never create template/config files next to an operator's original source.
    with tempfile.TemporaryDirectory(prefix="omrchecker-adapter-") as directory:
        work = Path(directory)
        shutil.copyfile(image_path, work / "sheet.png")
        for resource in ("marker.png", "reference.png"):
            source = image_path.with_name(resource)
            if source.is_file():
                shutil.copyfile(source, work / resource)
        return _read_sheet(root, work / "sheet.png", payload)


def _read_sheet(root: Path, image_path: Path, payload: dict) -> dict:
    if not (root / "src/core.py").is_file():
        raise InputError("OMRChecker engine files are missing. Restore omr/engine or check OMRCHECKER_ROOT.")
    try:
        revision = (root / "REVISION").read_text(encoding="utf8").strip()
    except OSError as exc:
        raise InputError("The bundled OMRChecker revision is missing. Restore omr/engine/REVISION.") from exc
    if not re.fullmatch(r"[0-9a-f]{40}", revision):
        raise InputError("The bundled OMRChecker revision is invalid. Restore omr/engine/REVISION.")
    template_data = bounded_template(payload.get("template"), image_path.with_name("marker.png").is_file(), image_path.with_name("reference.png").is_file())
    config_data = payload.get("config", {})
    if not isinstance(config_data, dict):
        raise InputError("Optional tuning config must be a JSON object.")
    config_data = deepcopy(config_data)
    config_data["outputs"] = {"show_image_level": 0, "save_image_level": 0, "save_detections": False, "filter_out_multimarked_files": False}
    dimensions = config_data.setdefault("dimensions", {})
    if not isinstance(dimensions, dict) or any(type(v) is not int or not 100 <= v <= 3000 for v in dimensions.values()):
        raise InputError("Processing/display dimensions must be between 100 and 3000 pixels.")
    alignment = config_data.get("alignment_params", {})
    if not isinstance(alignment, dict) or not number_between(alignment.get("max_steps", 20),1,200,integer=True):
        raise InputError("Alignment must use at most 200 steps.")
    sys.path.insert(0, str(root))
    import cv2
    import screeninfo

    # Upstream initializes display metrics even when all window output is disabled.
    monitors = screeninfo.get_monitors
    try:
        screeninfo.get_monitors = lambda: [SimpleNamespace(width=1920, height=1080)]
        from src.template import Template
        from src.utils.parsing import get_concatenated_response, open_config_with_defaults
    finally:
        screeninfo.get_monitors = monitors

    cv2.setNumThreads(1)
    template_path, config_path = image_path.with_name("template.json"), image_path.with_name("config.json")
    template_path.write_text(json.dumps(template_data), encoding="utf8")
    config_path.write_text(json.dumps(config_data), encoding="utf8")
    try:
        config = open_config_with_defaults(config_path)
        template = Template(template_path, config)
    except (Exception, SystemExit) as exc:
        raise InputError("OMRChecker rejected the template or tuning config. Check its native schema, field labels and preprocessing options.") from exc
    # Validate expanded geometry before any upstream array slices or processing.
    total = 0
    for block in template.field_blocks:
        width, height = block.bubble_dimensions
        if width < 2 or height < 2 or width > 512 or height > 512:
            raise InputError("Bubble dimensions must be between 2 and 512 pixels.")
        for bubbles in block.traverse_bubbles:
            total += len(bubbles)
            for bubble in bubbles:
                if bubble.x < 0 or bubble.y < 0 or bubble.x + width > template.page_dimensions[0] or bubble.y + height > template.page_dimensions[1]:
                    raise InputError("A bubble box leaves the template page. Correct the template dimensions and positions.")
    if total > 10_000:
        raise InputError("Template exceeds 10000 bubble boxes.")
    image = cv2.imread(str(image_path), cv2.IMREAD_GRAYSCALE)
    if image is None or image.size > 8_000_000:
        raise InputError("Use a decodable PNG/JPEG sheet of at most 8 megapixels.")
    processed = template.image_instance_ops.apply_preprocessors(image_path, image, template)
    if processed is None:
        raise InputError("OMRChecker could not find the page or markers. Check the template and source image.")
    responses, marked, multi, _ = template.image_instance_ops.read_omr_response(template, processed, "sheet.png")
    identity = read_identity(cv2.resize(processed, tuple(template.page_dimensions)), payload.get("identity_regions") or [], image_path.parent)
    fields = []
    for block in template.field_blocks:
        for bubbles in block.traverse_bubbles:
            label = bubbles[0].field_label
            raw = responses[label]
            values = [] if raw == block.empty_val else decode_values(raw, [b.field_value for b in bubbles])
            uncertain = values is None or len(values) != 1
            reason = "Unclear concatenated response; inspect the overlay." if values is None else "No marked bubble; confirm blank." if not values else "Multiple marked bubbles; review this field." if len(values) > 1 else ""
            fields.append({"label": label, "value": raw, "selection": values or [], "uncertain": uncertain, "reason": reason})
    scale = min(1, 1600 / max(marked.shape[:2]))
    preview = cv2.resize(marked, None, fx=scale, fy=scale) if scale < 1 else marked
    ok, encoded = cv2.imencode(".jpg", preview, [cv2.IMWRITE_JPEG_QUALITY, 82])
    if not ok or len(encoded) > 2_500_000:
        raise InputError("Could not produce a bounded detection overlay.")
    return {"engine": "OMRChecker", "revision": revision, "responses": get_concatenated_response(responses, template), "fields": fields, "multiple_marked": bool(multi), "overlay": base64.b64encode(encoded).decode("ascii"),"identity":identity}


def main() -> None:
    parser = argparse.ArgumentParser()
    parser.add_argument("--root", type=Path, required=True)
    parser.add_argument("--image", type=Path, required=True)
    args = parser.parse_args()
    try:
        payload = json.loads(sys.stdin.read(400_000))
        # Upstream logs and optional displays must never corrupt the JSON protocol.
        with open(os.devnull, "w") as sink, redirect_stdout(sink), redirect_stderr(sink):
            result = read_sheet(args.root.resolve(), args.image.resolve(), payload)
        print(json.dumps(result, allow_nan=False))
    except InputError as exc:
        print(json.dumps({"error": {"code": "invalid_omr_template", "message": str(exc)}}))
        sys.exit(1)
    except Exception:
        print(json.dumps({"error": {"code": "omrchecker_failed", "message": "OMRChecker could not process this sheet. Check its template, marker/reference images and installed runtime."}}))
        sys.exit(1)


if __name__ == "__main__":
    main()
