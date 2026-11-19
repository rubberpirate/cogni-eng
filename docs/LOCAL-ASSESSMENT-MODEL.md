# Local assessment model

For the inspected machine (RTX 4060 Laptop, 8 GB VRAM; 24 GB system RAM), the selected starting candidate is **Qwen3.5-9B Q4_K_M**, served locally by Ollama. The [official Ollama tag](https://ollama.com/library/qwen3.5:9b-q4_K_M) lists a 6.6 GB package and Apache-2.0 license. The [upstream Qwen model card](https://huggingface.co/Qwen/Qwen3.5-9B) describes its multimodal capabilities. This is a hardware-aware engineering recommendation, not evidence that the model generates correct JEE/NEET papers.

Use the CPU Tesseract OCR engine first, then this same local language model to structure question text/options and explicit printed keys, or to generate original exam/subject/topic/difficulty groups. Go validates the structure and marking rules and retains the source, job progress and reviewed draft. No paid model is used as a fallback for generation or document normalization. Existing hosted adapters remain only for older direct contract tests/other separately configured features; assessment model selection returns the local adapter. An explicitly labelled development fixture remains permitted in development/test environments.

## Setup

Install Ollama from [its official download instructions](https://ollama.com/download/linux), then download the explicit quantization tag:

```sh
ollama pull qwen3.5:9b-q4_K_M
ollama serve
```

A service-managed Ollama installation may already be serving; do not start a second server on the same port. Configure the Go worker as in `.env.example`:

```dotenv
ACADEMIC_MODEL_PROVIDER=local
LOCAL_MODEL_URL=http://127.0.0.1:11434
LOCAL_MODEL_NAME=qwen3.5:9b-q4_K_M
OCR_PROVIDER=tesseract
```

Restart the worker after changing these values. The initial software/model download needs internet access; subsequent assessment inference is local. The adapter accepts only loopback HTTP endpoints and does not follow redirects. The application sends curriculum or OCR text without attaching student rosters or profile records. The [Ollama chat API](https://docs.ollama.com/api/chat) is used with `stream:false`, `think:false`, JSON output, a bounded 16K context and 8,192 output-token limit. [Structured-output documentation](https://docs.ollama.com/capabilities/structured-outputs) explains JSON mode; Go independently validates the result and rejects truncated responses, malformed fields or wrong generation counts.

The 6.6 GB download size is not a claim of total VRAM consumption. Runtime buffers and context need additional memory; partial CPU offload may be necessary on an 8 GB GPU. Do not keep the existing LightOn vision model resident alongside it. The new upload wizard explicitly uses CPU Tesseract, and local requests use `keep_alive:0` to release the language model after completion. Begin with groups of 3–5 questions, then measure before increasing counts. A larger [27B model](https://huggingface.co/Qwen/Qwen3.5-27B) is not the default for this machine; its quantized weights and runtime overhead need substantially more memory and may be slow on CPU. A smaller 4B variant is a speed/memory alternative, not an established quality substitute.

## Document processing and recovery

OCR retains the original transcription before local AI processing. Numbered question boundaries divide long papers into groups of at most eight questions; no group is silently truncated. Unstructured large text or oversized groups produce a recoverable error asking for a split/correction. Completed AI chunks are retained on the durable job and reused after retry, keyed by source-text/model hash. Go reports processing stages, rechecks scope/cancellation before calls and saves, and retains source pages when inference fails. Requests have a ten-minute timeout within a fifteen-minute worker bound.

The model is instructed to retain keys only when explicitly supplied or marked. Missing answers remain editable and block publication. Separate answer-key files use the existing reviewed key-import flow; the simplified editor uses CPU OCR for these files too; source IDs preserve key matching after draft questions are reordered. Missing worked solutions do not cause an invented scientific explanation: publication can supply truthful marking guidance based on the faculty-approved answer and marks. Faculty must review wording, keys and scoring before publication. Published versions and existing sittings remain immutable.

## Evidence and qualification

Ollama and the Qwen weights were **not installed or benchmarked** in this increment. The adapter and OCR-to-model pipeline passed local HTTP contract fixtures, not live-model quality evaluation. No JEE/NEET accuracy percentage or generation-speed claim is made.

Before faculty use, evaluate an original/licensed holdout across physics, chemistry, mathematics and biology: question/option/number completeness; equation/unit fidelity; explicit-key precision; single/multiple/numerical answer validity; topic/range coverage; distractor quality; independently checked worked solutions; difficulty agreement; latency; RAM/VRAM; and retry/cancellation behavior. Compare against faculty-authored keys without showing those keys to generation. Record subject-level failures and retain human review. Published general benchmark results cannot replace this application-specific evaluation.

There are no per-request hosted API fees in this path; the device, storage and electricity still have costs.
