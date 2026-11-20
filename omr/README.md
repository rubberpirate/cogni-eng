# Native OMRChecker adapter

This adapter imports the bundled [OMRChecker](engine/UPSTREAM.md) engine in
`omr/engine` and calls its `Template`, configured preprocessors and
`ImageInstanceOps.read_omr_response`. It does not reimplement the detector or run
the upstream grading/evaluation module. Go retains authentication, staff/course
scope, validation and academic grading authority. No separate checkout, Git
submodule or OMR service is required.

Bundled upstream revision: `5cf44a5a9c7a49e2e541c6e19aefd09a41a4d494`.
Responses report that revision from `engine/REVISION`; runtime does not need Git.
Upstream runtime source, the MIT license and contributor notices are included in
`engine/`. Upstream tests/sample documents and supplied private examination data
are excluded. See [the preserved license](engine/LICENSE).

## Setup and use

From a fresh project checkout, install `uv` and run from the application root:

```sh
bun run omr:setup
bun run dev
```

For backend-only hosting, `uv sync --project omr --frozen` installs the same
dependencies without requiring Bun. Install Tesseract and its English language
data for identity OCR, and Poppler for PDF scan intake. The engine source is
already included; setup downloads Python dependencies, not another repository.

`omr/uv.lock` freezes the CPU runtime, including headless OpenCV. No Ollama/GPU or
hosted API is used for this reader. Normally leave these optional overrides
empty. If your project/runtime is elsewhere, paths refer to this project:

```dotenv
OMRCHECKER_ROOT=/path/to/cogni-fresh/omr/engine
OMRCHECKER_PYTHON=/path/to/cogni-fresh/omr/.venv/bin/python
OMRCHECKER_ADAPTER=/path/to/cogni-fresh/omr/bridge.py
```

Defaults are `omr/engine`, `omr/.venv/bin/python` and `omr/bridge.py` relative to
the application working directory. If the adapter path is overridden, the
default engine lives in its adjacent `engine/` directory. The backend Docker
image includes Python, the frozen dependencies, adapter and bundled engine; no
external checkout mount is needed. For personal-server PM2 hosting see
[SELF-HOSTING.md](../docs/SELF-HOSTING.md).

Open **Upload OMR answer sheets**. There is one scanning flow:

1. Select the published **Assessment**. Its pinned questions, key, marks, deductions and section policies determine the grade.
2. Choose a named, saved **OMR template** from the institution's shared library. Templates describe sheet geometry, not assessment keys or marks. A template needs at least as many answer rows as the assessment, with matching option counts on the rows being used. Extra rows are ignored; for example, a 150-row sheet can serve a 20-question assessment using its first 20 rows. The last selected template is remembered for that assessment.
3. Upload one image, several files together, or a multi-page PDF and choose **Check answer sheets**. Up to 20 files can be submitted at once. Each upload is stored privately and queued through the existing Go/River worker, with progress, cancellation and retry. A repeated file reuses its existing check rather than creating duplicate grades.

Results list registration number, optional student name, marks, correct/wrong question numbers and review status. Registration numbers remain strings, preserving leading zeros. No learner account, enrollment or assignment is created or required. Draft papers may now be published without online recipients; selected recipients are still assigned when present. The **Results** page includes these registration-based grades; existing account-linked submissions remain separate.

Uncertain identity, blank/unclear response readings and invalid single-choice multiple marks remain provisional. Open **Review sheet**, check the private marked image, correct identity/answers and save with explicit confirmation and a reason. Go uses `academic.Mark` for both initial and reviewed marks. Previous answers and grades remain in revision evidence. A duplicate registration within an assessment cannot overwrite an existing graded sheet; a duplicate/wrong page can be explicitly excluded instead.

### Create a template

**Create an OMR template** sits below the scanner and results. No assessment selection is required. Upload a blank or sample image and draw frames. Label each as **Registration number**, **Student name** or **Answers**. Identity fields are optional and can be added in any order. New frames contain no generated circles. For an answer column, choose numbered or lettered options and **Mark first two rows**. The selected column enlarges; click each bubble in the first two rows, check the inferred count and save the frame. Repeat for other columns. Whole-sheet fit, an expanded native dialog, keyboard controls and optional bubble guides are available.

Choose a faculty-defined name and **Save template**. The template, tuning and generated reference are saved to the server and become a selectable template for future assessments across teaching batches in the institution. Existing templates are retained. The creating course/campus remains storage provenance for private reference-image permissions; it does not restrict which assessment can use the geometry. Migration `00010_saved_omr_checks.sql` adds template, upload and account-free sheet-grade persistence. Reference/source/marked images use the existing private object store. A portable `.omr.json` export and separate generated resources remain available at the bottom; local files are no longer the primary persistence mechanism.

The reference masks selected identity boxes and neutralizes configured bubble interiors. Native `FeatureBasedAlignment` uses it, so a separate marker upload is unnecessary. Blank sources are preferable: marks outside selected regions are not guaranteed to be removed. The visual editor still supports importing older portable profiles; they are saved through the same server validation. Drawing uses cached geometry and a transient SVG outline updated on animation frames, without rebuilding the field list on pointer moves.

### Local recognition and API

Names/registration snippets are read locally by Tesseract (`tesseract` and English language data must be installed). Missing or low-confidence text stays available for manual review. Handwritten identity accuracy is unqualified. No automatic account/roster matching is performed.

The saved flow uses scoped, session/CSRF-protected generated contracts:

- `GET/POST education/omr-templates` lists/saves named templates independently of an assessment. Listing requires active academic staff access and returns institution-wide metadata, with no reference-image bytes or private asset identifiers. Multipart saves contain an authorized `course_id` for storage provenance, `definition` JSON and reference image `file`; no `paper_id` is required.
- `GET/POST education/omr-checks` lists/queues checks; uploads contain `paper_id`, `template_id`, `file`. Results use stable `before` cursors, 100 uploads per page.
- `GET education/omr-checks/{id}` and `GET education/omr-sheets/{id}` retrieve authorized upload/sheet evidence.
- `POST education/omr-sheets/{id}/review` applies versioned identity/answer review or exclusion.

Template creation requires active teaching access in its storage course; library listing enforces active staff, institution and MFA checks. Checking/review operations use the selected paper's full staff/subject scope. Shared template metadata never grants access to the creating course's private reference image. Private marked images require the assessment's paper responsibilities. Original native-preview and registered-evidence APIs remain for compatibility/regression checks; raw-template upload, pixel diagnostics and the registered scanner are not alternative faculty UI choices.

## Execution boundaries

Selection-grid templates currently support 1–200 consecutive questions with 2–8 options. Only the first N rows are used for an N-question published assessment; those rows must match its option counts. Numerical digit grids and written responses need additional field configuration and are rejected rather than silently graded as selection questions. The underlying original preview remains broader, but is not a separate scanner option.

Images are capped at 8 MiB/8 megapixels; source PDFs at 50 MiB and 20 pages. PDF pages are rasterized locally through Poppler. Template/tuning JSON limits remain 256/64 KiB. Native execution is headless, capped at forty seconds per page and six MiB output, with a minimal environment excluding application/provider credentials. Template resource paths are rewritten to uploaded/private resources. Original sheets and the installed checkout are unchanged. Native errors retain upload/job evidence for recovery, without a silent detector fallback.

The adapter supports installed native `CropPage`, `CropOnMarkers`, `FeatureBasedAlignment`, `Levels`, `MedianBlur` and `GaussianBlur` processors. Visual templates use reference alignment. The old generic `auto_align` is disabled by default. General geometry discovery, handwriting/lighting/erasure diversity, all preprocessing options, production/container behavior and full accessibility acceptance remain unqualified.

## Original test kit

The app provides [a filled sheet](../apps/web/public/samples/omrchecker/filled-sheet.png), [its native template](../apps/web/public/samples/omrchecker/template.json) and [expected filled responses](../apps/web/public/samples/omrchecker/expected-answers.json). They are original synthetic CC0 test material generated by `make_sample.py`. The roll is `123456`; question 7 is blank and question 12 contains A+C. These are marked responses, not an answer key for awarding grades.

Run unit and real-engine checks:

```sh
omr/.venv/bin/python -m unittest discover -s omr -p 'test_*.py'
RUN_OMRCHECKER_TESTS=1 go test ./internal/scholar -run TestBundledOMRChecker
RUN_OMRCHECKER_TESTS=1 bunx playwright test tests/browser/omrchecker.spec.ts
RUN_OMRCHECKER_TESTS=1 bunx playwright test tests/browser/omr-visual-setup.spec.ts
bun test tests/omr-template.test.ts
```

The Python suite always exercises the bundled engine on the original filled
sample, including its upstream revision and absence of a Git subprocess. The
Go check runs with an empty home and PATH and no engine path overrides. The
opt-in browser test uses the bundled engine and actual Go upload route. No HTTP
recognition response is mocked. Field-scan accuracy, erasures, lighting/layout
diversity and handwritten identity recognition remain unqualified.
