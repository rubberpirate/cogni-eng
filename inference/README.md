# Private vision inference runtime

This process performs model inference only. The Go worker owns file authorization, durable jobs, interpretation, review and marking. It has no institution database or application session access.

Run from the repository root:

```sh
bun run vision --warm lightonai/LightOnOCR-2-1B
```

The first run creates the frozen Python environment; later starts reuse it and verify the pinned direct dependency versions. `bun run vision --install` explicitly resynchronizes it. The setup downloads model weights into ignored `.local/models`. It generates a private `VISION_API_KEY` in the existing ignored `.env` if absent. Restart `bun run dev` after configuration changes. Keep the key on the worker and inference service only. The service binds loopback port 58010 and accepts authenticated inline PNG/JPEG requests; it never fetches image URLs or accepts filesystem paths from the request.

Set `VISION_AUTOSTART=1` for `bun run dev` to manage the installed runtime alongside the Go worker. Stop a separately started runtime first to avoid a port conflict.

Select **LightOnOCR-2-1B** in the upload form to request the model explicitly. `OCR_PROVIDER=lightonocr` changes the default development worker choice. `OCR_PROVIDER=tesseract` keeps the bounded native OCR path. There is no silent fallback between engines. A missing/unreachable model service retains the uploaded source and recoverable job. Key uploads have the same explicit engine selection.

Under **Scan & check → Read an existing OMR format**, calibrate the image, then select the local vision comparison. Pixel measurements remain the observed baseline; classifier outputs are labelled proposals and conflicts require review. The model has no answer key, grade policy or learner identity. No diagnostic output creates marks. Jobs persist the source, calibration, actual model/revision and acknowledged band outputs; retry reuses completed bands. Generic field accuracy is not established by provider benchmarks.

The original bubble-mark classifier runs on CPU. The runtime loads one language/vision model at a time to accommodate consumer GPUs; CPU mark classification has a separate bounded execution slot. Loading another model unloads the previous weights. Requests while busy receive 503 and remain recoverable in Go. Images are limited to 8 MiB/8 megapixels; output to 4096 tokens; generation has a 120-second stopping limit after model loading. Go adds a request timeout and job cancellation checks. GPU execution may finish its current bounded step after a client disconnects; cancellation does not claim instantaneous GPU interruption. No persistent prompt history or shared document cache is used.

Pinned models (third-party Apache-2.0 weights are downloaded separately; original CC0-1.0 classifier weights are bundled):

- `lightonai/LightOnOCR-2-1B` at `be7c41313c30e940fe5571de59e5fabc779e00c8` — image-only document transcription.
- `learning/bubble-marks-v1` — original small convolutional mark classifier, bundled as Safetensors. Training used 10,000 newly authored synthetic patches; the separate 2,000-patch synthetic holdout had 1,999 correct classifications. This is not field validation or a calibrated probability. See its training report and MODEL-NOTICE.md.
- `Qwen/Qwen3.5-2B` at `15852e8c16360a2fea060d615a32b45270f8a8fc` — experimental general vision comparison. It was unreliable on the supplied OMR image and is not the default. `OMR_MODEL_PROVIDER=qwen_vision` is for explicit evaluation; the default is `bubble_classifier`.

Official documentation: [LightOn model card](https://huggingface.co/lightonai/LightOnOCR-2-1B), [LightOn serving/benchmark repository](https://github.com/lightonai/LightOnOCR), [Qwen model card](https://huggingface.co/Qwen/Qwen3.5-2B). Published benchmark claims do not establish accuracy for this application's papers or OMR.

`uv sync --project inference --frozen` reproduces dependencies. The Dockerfile and the optional `vision` profile in `compose.application.yaml` provide a GPU-container path; configure the worker URL as `http://vision:58010/v1`, share the private key, and install the NVIDIA container runtime on the host. Container build/run qualification is separate from host tests. Do not publish the inference port publicly.

Protocol guard checks:

```sh
uv run --project inference --frozen --no-sync python -m unittest discover -s inference
```

GPU availability, educational quality, real-image findings and deployment qualification must be reported in `docs/TEST-REPORT.md`; a passing HTTP fixture is not live model evidence.
