# Native OMRChecker adapter

This adapter imports the operator's real [OMRChecker](https://github.com/Udayraj123/OMRChecker) checkout and calls its `Template`, configured preprocessors and `ImageInstanceOps.read_omr_response`. It does not reimplement the detector or run the upstream grading/evaluation module. Go retains authentication, staff/course scope, validation and academic grading authority.

Tested upstream revision: `5cf44a5a9c7a49e2e541c6e19aefd09a41a4d494`. Responses include the actual checkout revision. Upstream is MIT licensed, copyright 2024–present Udayraj Deshmukh and other contributors; its source and license remain in the separately installed checkout. No upstream source or supplied private examination data is vendored here.

## Setup and use

The current machine already has `~/OMRChecker`. From the application root:

```sh
bun run omr:setup
bun run dev
```

`omr/uv.lock` freezes the CPU runtime, including headless OpenCV. No Ollama/GPU or hosted API is used for this reader. If your checkout or runtime is elsewhere, set these before starting Go:

```dotenv
OMRCHECKER_ROOT=/path/to/OMRChecker
OMRCHECKER_PYTHON=/path/to/cogni-fresh/omr/.venv/bin/python
OMRCHECKER_ADAPTER=/path/to/cogni-fresh/omr/bridge.py
```

Defaults are `~/OMRChecker`, `omr/.venv/bin/python` and `omr/bridge.py` relative to the application working directory. Containers need the checkout/runtime mounted or installed explicitly; this increment does not change the container image.

Open **Upload OMR answer sheets → Read with OMRChecker**. Choose the permitted batch and a PNG/JPEG sheet, preferably blank when configuring a new layout. The default visual setup has three steps:

1. Draw boxes around the student name and roll number (optional).
2. Draw one answer-column box. Click every bubble left to right in the first row, then the row immediately below it: eight clicks for four options. Remaining rows fill automatically. Check the count and repeat for each column. Blocks can be edited or removed, and bubble size can be adjusted.
3. Name and review the layout, then choose **Use this layout**. Template JSON, tuning and a reference image are created together and passed directly to the native reader. No extra file selection is needed.

The reference masks the selected identity boxes and neutralizes the configured bubble interiors. Native `FeatureBasedAlignment` uses this reference, so a separate marker image is unnecessary for this setup. A blank source is preferable; generated references from filled sheets are not guaranteed to remove ink outside the selected areas. Save a reusable `.omr.json` file to keep the layout across reloads; choose **Load saved layout** on later visits. Uploading another sheet keeps an applied layout. Separate native files are downloadable under **Generated OMRChecker files**. Profiles are portable browser downloads, not stored institution templates.

Keyboard users can focus the sheet and use arrows plus Enter to place corners/circles (Shift: 50 pixels, Alt: one pixel, otherwise five pixels; Escape cancels). The canvas supports zoom. Layouts support 2–8 options, up to eight evenly spaced answer blocks and 500 questions total, with up to 200 per block. Check the generated circles against the whole column before reading; irregular spacing needs separate blocks.

**Use existing files** retains native template JSON upload, with optional tuning JSON, marker image for `CropOnMarkers`, and blank reference image for `FeatureBasedAlignment`. The returned table contains original detected fields and combined template fields (such as a bubbled roll number), plus the aligned detection overlay and a JSON download. Blank, multiple or ambiguously concatenated responses require review.

Selected identity snippets are read locally from the aligned sheet with Tesseract (`tesseract` plus English language data must be installed). Detected text is editable, with low-confidence values flagged. OCR failure leaves manual identity entry available without discarding bubble responses. These text readings do not establish a roster match, and handwriting accuracy is unqualified.

The authenticated multipart endpoint is `POST /api/v1/institutions/{org}/education/omrchecker-preview`, with `course_id`, `file`, `template`, optional `config`, `marker`, `reference` and `identity_regions` (JSON array of unique `student_name` / `roll_number` kinds with normalized `rect` coordinates). Go validates at most two in-bounds identity regions. It uses the existing session and CSRF boundary and generated OpenAPI/TypeScript contracts. It creates no result, assignment, roster match or final mark.

## Execution boundaries

Each PNG/JPEG is limited to 8 MiB/8 megapixels; template JSON to 256 KiB and tuning JSON to 64 KiB. Expanded templates allow at most 500 fields and 10000 bubbles. Image dimensions, field-range expansion, kernels, marker scale searches and plugin names are bounded before upstream processing. Template-supplied marker/reference paths are replaced with the separately uploaded resources; they cannot select host files.

Execution is headless, capped at forty seconds and six MiB of structured output, with a minimal subprocess environment that excludes application/provider credentials. All temporary images/configs are private and removed on completion/failure. Original uploads and the installed checkout are not modified. The preview itself is transient; selected browser files remain available to retry. It never silently substitutes the old pixel reader if OMRChecker fails.

The adapter supports the installed native `CropPage`, `CropOnMarkers`, `FeatureBasedAlignment`, `Levels`, `MedianBlur` and `GaussianBlur` processors. Default upstream tuning disables the deprecated generic `auto_align`; uploaded compatible tuning remains explicit. A template is required, either generated visually or supplied. Arbitrary-layout discovery, batch reconciliation, PDF intake for this endpoint and automatic native-template grading remain outstanding. Existing registered OMR grading remains available separately; its advanced pixel-reader diagnostics are collapsed by default.

## Original test kit

The app provides [a filled sheet](../apps/web/public/samples/omrchecker/filled-sheet.png), [its native template](../apps/web/public/samples/omrchecker/template.json) and [expected filled responses](../apps/web/public/samples/omrchecker/expected-answers.json). They are original synthetic CC0 test material generated by `make_sample.py`. The roll is `123456`; question 7 is blank and question 12 contains A+C. These are marked responses, not an answer key for awarding grades.

Run unit and real-engine checks:

```sh
omr/.venv/bin/python -m unittest discover -s omr -p 'test_*.py'
RUN_OMRCHECKER_TESTS=1 omr/.venv/bin/python -m unittest discover -s omr -p 'test_*.py'
RUN_OMRCHECKER_TESTS=1 bunx playwright test tests/browser/omrchecker.spec.ts
RUN_OMRCHECKER_TESTS=1 bunx playwright test tests/browser/omr-visual-setup.spec.ts
bun test tests/omr-template.test.ts
```

The opt-in browser test uses the installed engine and actual Go upload route. No HTTP recognition response is mocked. Field-scan accuracy, erasures, lighting/layout diversity and handwritten identity recognition remain unqualified.
