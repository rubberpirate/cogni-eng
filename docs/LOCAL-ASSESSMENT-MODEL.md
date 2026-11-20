# Local assessment model

For the inspected machine (RTX 4060 Laptop, 8 GB VRAM; 24 GB system RAM), the selected starting candidate is **Qwen3.5-9B Q4_K_M**, served locally by Ollama. The [official Ollama tag](https://ollama.com/library/qwen3.5:9b-q4_K_M) lists a 6.6 GB package and Apache-2.0 license. The [upstream Qwen model card](https://huggingface.co/Qwen/Qwen3.5-9B) describes its multimodal capabilities. This is a hardware-aware engineering recommendation, not evidence that the model generates correct JEE/NEET papers.

Use a usable embedded PDF text layer first, or CPU Tesseract for scanned PDFs/images, then this same local language model to structure question text/options and explicit printed keys, or to generate original exam/subject/topic/difficulty groups. Go validates the structure and marking rules and retains the source, job progress and reviewed draft. No paid model is used as a fallback for generation or document normalization. Existing hosted adapters remain only for older direct contract tests/other separately configured features; assessment model selection returns the local adapter. An explicitly labelled development fixture remains permitted in development/test environments.

## Setup

Install Ollama from [its official download instructions](https://ollama.com/download/linux), then download the explicit quantization tag:

```sh
ollama pull qwen3.5:9b-q4_K_M
ollama serve
```

A service-managed Ollama installation may already be serving; do not start a second server on the same port. Alternatively, `bun run model:local` reuses an existing loopback server or starts the installed Ollama with cloud disabled, one inference slot and a minimal environment that excludes application credentials. Install the appropriate GPU backend for the operating system and check that `nvidia-smi` and Ollama detect the GPU. Configure the Go worker as in `.env.example`:

```dotenv
ACADEMIC_MODEL_PROVIDER=local
LOCAL_MODEL_URL=http://127.0.0.1:11434
LOCAL_MODEL_NAME=qwen3.5:9b-q4_K_M
OCR_PROVIDER=tesseract
```

Restart the worker after changing these values. The initial software/model download needs internet access; subsequent assessment inference is local. The adapter accepts only loopback HTTP endpoints and does not follow redirects. The application sends curriculum or OCR text without attaching student rosters or profile records. The [Ollama chat API](https://docs.ollama.com/api/chat) is used with `stream:false`, `think:false`, JSON output, a bounded 16K context and 8,192 output-token limit. [Structured-output documentation](https://docs.ollama.com/capabilities/structured-outputs) explains JSON mode; Go independently validates the result and rejects truncated responses, malformed fields or wrong generation counts.

The 6.6 GB download size is not a claim of total VRAM consumption. Runtime buffers and context need additional memory; partial CPU offload may be necessary on an 8 GB GPU. Do not keep the existing LightOn vision model resident alongside it. The upload wizard uses CPU OCR for scans, and local requests use `keep_alive:"2m"` to avoid reloading between chunks. Calls are serialized within each Go worker; the launch helper also requests one Ollama inference slot. Begin with groups of 3–5 questions, then measure before increasing counts. A larger [27B model](https://huggingface.co/Qwen/Qwen3.5-27B) is not the default for this machine; its quantized weights and runtime overhead need substantially more memory and may be slow on CPU. A smaller 4B variant is a speed/memory alternative, not an established quality substitute.

## Document processing and recovery

Document processing retains the original transcription before local AI processing. Digital PDFs use Poppler's text layer when usable, preserving printed answer lines that raster OCR can omit; image-only PDFs fall back to OCR. Both paths enforce the 1–50-page limit. Numbered question boundaries, including Q.1/Q2 labels, divide long papers into groups of at most eight questions; no group is silently truncated. Unstructured large text or oversized groups produce a recoverable error asking for a split/correction. Completed AI chunks are retained on the durable job and reused after retry, keyed by source-text/model hash. Go reports processing stages, rechecks scope/cancellation before calls and saves, and retains source pages when inference fails. Requests have a ten-minute timeout within a fifteen-minute worker bound. Extraction requests constrain wire types with a JSON schema; Go remains responsible for business validation. For recognized numbered sections, a final source pass checks question count/order and restores explicit single/multiple/numerical types, full marks and penalty magnitudes across chunk boundaries and subject-number resets. Numerical source sections without supplied keys remain unanswered.

The model is instructed to retain keys only when explicitly supplied or marked. Go additionally reconciles literal `Answer:` lines by printed question identity, including multiple selections, decimal values and one accepted range; an answer alone is not a worked solution. Ambiguous numbering or unsupported printed rules remain flagged for review. Disjoint accepted ranges are not representable by the current numerical grading rule and must not be flattened into one broader range. Missing answers remain editable and block publication. Separate answer-key files use the existing reviewed key-import flow; the simplified editor uses CPU OCR for these files too; source IDs preserve key matching after draft questions are reordered. Missing worked solutions do not cause an invented scientific explanation: publication can supply truthful marking guidance based on the faculty-approved answer and marks. Faculty must review wording, keys and scoring before publication. Published versions and existing sittings remain immutable.

## Evidence and qualification

The installed Qwen3.5-9B Q4_K_M completed real connected generation: two JEE groups (three mathematics and two physics questions, 2.8 minutes) and three NEET biology questions (59.6 seconds). The adapter now performs a second local consistency review and retains the initial draft for audit after an earlier run returned a key contradicting its own solution. All eight final selected answers were independently checked; one physics solution still needed an intermediate-velocity correction through the faculty editor before publishing the five-question, 20-mark paper to a synthetic learner. Automated review does not establish scientific correctness or exam difficulty.

After GPU recovery on 2026-11-20, the system Ollama CUDA backend offloaded all 33 layers to the RTX 4060, with approximately 6.8 GB VRAM observed during extraction. These are individual functional observations, not a throughput benchmark. Real full-paper tests establish counts/order, source types/keys/marks and reload/retry behavior on the three supplied PDFs (48/51/51 questions). Per-chunk count/order checks prevent incomplete model output from becoming a reusable successful chunk; retry regenerates invalid old chunks selectively. These functional checks do not establish equation/diagram fidelity. See TEST-REPORT.md for completed runs and limitations.

An observed extraction defect remains: the 2025 paper's mathematics question 3 lost powers and the reciprocal argument in a two-dimensional formula. A usable PDF text layer and complete question count do not prove mathematical fidelity. Compare formulas and diagrams with the retained original PDF and correct them in the editor before publishing. This text pipeline is not qualified for unattended import of complex examination papers.

Before faculty use, evaluate an original/licensed holdout across physics, chemistry, mathematics and biology: question/option/number completeness; equation/unit fidelity; explicit-key precision; single/multiple/numerical answer validity; topic/range coverage; distractor quality; independently checked worked solutions; difficulty agreement; latency; RAM/VRAM; and retry/cancellation behavior. Compare against faculty-authored keys without showing those keys to generation. Record subject-level failures and retain human review. Published general benchmark results cannot replace this application-specific evaluation.

There are no per-request hosted API fees in this path; the device, storage and electricity still have costs.
