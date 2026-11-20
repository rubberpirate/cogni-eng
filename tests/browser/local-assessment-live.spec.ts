import { test, expect, type Page } from "@playwright/test";
import { readFileSync, writeFileSync, mkdirSync } from "node:fs";
import { resolve } from "node:path";
import { signInStaff } from "./staff-sign-in";

test.skip(
  process.env.RUN_LOCAL_MODEL_TESTS !== "1",
  "Opt-in: real local inference and owner-supplied papers, no HTTP model fixture.",
);
const source = resolve(process.env.LOCAL_TEST_PAPERS ?? "test-papers");
const artifacts = ".local/live-assessment";
const fixture = JSON.parse(
  readFileSync(".local/academic-browser-access.json", "utf8"),
);
async function setup(
  page: Page,
  method: string,
  subject: string,
  exam = "IIT-JEE",
) {
  await signInStaff(page);
  await page
    .getByRole("button", { name: new RegExp(method) })
    .filter({ has: page.locator("strong") })
    .click();
  await page
    .getByLabel("Assessment name", { exact: true })
    .fill(`Live ${method} ${Date.now()}`);
  await page
    .getByLabel("Student batch", { exact: true })
    .selectOption(fixture.course);
  await page.getByLabel("Main subject", { exact: true }).fill(subject);
  await page
    .getByLabel("Main topic", { exact: true })
    .fill(`Live validation ${Date.now()}`);
  await page.getByLabel("Paper type", { exact: true }).selectOption(exam);
  await page
    .getByRole("button", { name: "Continue to questions", exact: true })
    .click();
  await expect(page.locator(".editor-toolbar")).toBeVisible();
  return `/api/v1/institutions/${fixture.org}/education/`;
}
function record(name: string, value: unknown) {
  mkdirSync(artifacts, { recursive: true });
  writeFileSync(`${artifacts}/${name}.json`, JSON.stringify(value, null, 2));
}
async function savedDraft(page: Page, base: string) {
  const id = new URLSearchParams(new URL(page.url()).hash.split("?")[1]).get(
    "draft",
  );
  expect(id).toBeTruthy();
  const response = await page.request.get(base + `drafts/${id}`);
  expect(response.ok()).toBe(true);
  return response.json();
}
async function waitForPaper(page: Page, base: string, expected?: number) {
  const began = Date.now();
  while (Date.now() - began < 14 * 60 * 1000) {
    const draft = await savedDraft(page, base);
    if (
      draft.spec.questions.length &&
      (!expected || draft.spec.questions.length === expected)
    )
      return draft;
    const ids = draft.spec.import_id
      ? (await (await page.request.get(base + "jobs")).json())
          .filter((j: any) => j.resource_id === draft.spec.import_id)
          .map((j: any) => j.id)
      : draft.spec.job_ids;
    for (const id of ids) {
      const j = await (await page.request.get(base + `jobs/${id}`)).json();
      if (["failed", "configuration_required", "cancelled"].includes(j.state)) {
        record(`failure-${id}`, j);
        throw Error(`${j.kind}: ${j.error}`);
      }
    }
    await page.waitForTimeout(4000);
  }
  throw Error(
    "Live paper processing exceeded fourteen minutes; retained sources/jobs can be retried.",
  );
}

test("real Ollama generates two IIT-JEE topic groups and retains editable questions", async ({
  page,
}) => {
  test.setTimeout(16 * 60 * 1000);
  const base = await setup(page, "Generate with AI", "Mathematics");
  await page.getByLabel("Group 1 subject", { exact: true }).fill("Mathematics");
  await page
    .getByLabel("Group 1 topic or topic range", { exact: true })
    .fill(
      "Quadratic equations with real coefficients; use explicit numerical coefficients and verify the roots.",
    );
  await page.getByLabel("Group 1 question count", { exact: true }).fill("3");
  await page
    .getByRole("button", { name: "Add another group", exact: true })
    .click();
  await page.getByLabel("Group 2 subject", { exact: true }).fill("Physics");
  await page
    .getByLabel("Group 2 topic or topic range", { exact: true })
    .fill(
      "One-dimensional motion with constant acceleration; explicit numerical data and SI units.",
    );
  await page.getByLabel("Group 2 question count", { exact: true }).fill("2");
  const began = Date.now();
  await page
    .getByRole("button", { name: "Generate question groups", exact: true })
    .click();
  const draft = await waitForPaper(page, base, 5);
  record("generated-jee", { elapsed_ms: Date.now() - began, draft });
  expect(new Set(draft.spec.questions.map((q: any) => q.stem)).size).toBe(5);
  for (const q of draft.spec.questions) {
    expect(q.options).toHaveLength(4);
    expect(q.key.length).toBeGreaterThan(0);
    expect(q.solution.length).toBeGreaterThan(10);
    expect(q.provenance).toContain("qwen3.5");
  }
  await page.reload();
  await expect(page.locator(".draft-question")).toHaveCount(5);
  await page.screenshot({
    path: `${artifacts}/generated-jee.png`,
    fullPage: true,
  });
});

test("real Ollama generates a NEET biology draft", async ({ page }) => {
  test.setTimeout(16 * 60 * 1000);
  const base = await setup(page, "Generate with AI", "Biology", "NEET");
  await page.getByLabel("Group 1 subject", { exact: true }).fill("Biology");
  await page
    .getByLabel("Group 1 topic or topic range", { exact: true })
    .fill("Cell organelles and their functions in eukaryotic cells");
  await page.getByLabel("Group 1 question count", { exact: true }).fill("3");
  const began = Date.now();
  await page
    .getByRole("button", { name: "Generate question groups", exact: true })
    .click();
  const draft = await waitForPaper(page, base, 3);
  record("generated-neet", { elapsed_ms: Date.now() - began, draft });
  expect(draft.spec.exam).toBe("NEET");
  await page.reload();
  await expect(page.locator(".draft-question")).toHaveCount(3);
  await page.screenshot({
    path: `${artifacts}/generated-neet.png`,
    fullPage: true,
  });
});

for (const [file, count] of [
  ["2025_1_English.pdf", 48],
  ["2024_1_English.pdf", 51],
  ["pdf.pdf", 51],
] as const) {
  test(`real full-paper upload and Ollama extraction: ${file}`, async ({
    page,
  }) => {
    test.setTimeout(18 * 60 * 1000);
    const began = Date.now();
    let base: string;
    if (
      process.env.RESUME_LOCAL_UPLOADS === "1" &&
      !process.env.FRESH_LOCAL_UPLOADS?.split(",").includes(file)
    ) {
      await signInStaff(page);
      const saved = JSON.parse(
        readFileSync(`${artifacts}/resume-uploads.json`, "utf8"),
      ).find((r: any) => r.filename === file);
      expect(saved).toBeTruthy();
      await page.goto(`/workspace#education/draft?draft=${saved.id}`);
      base = `/api/v1/institutions/${fixture.org}/education/`;
      await page
        .getByRole("button", { name: "Check processing / retry", exact: true })
        .click();
      await expect
        .poll(
          async () =>
            (
              await (
                await page.request.get(base + `jobs/${saved.job_id}`)
              ).json()
            ).state,
        )
        .not.toBe("cancelled");
    } else {
      base = await setup(page, "Upload question paper", "Mathematics");
      await page
        .getByLabel("Question paper file", { exact: true })
        .setInputFiles(resolve(source, file));
    }
    const draft = await waitForPaper(page, base);
    const imp = await (
      await page.request.get(base + `imports/${draft.spec.import_id}`)
    ).json();
    record(`upload-${file}`, {
      elapsed_ms: Date.now() - began,
      expected: count,
      draft,
      import: imp,
    });
    expect(draft.spec.questions).toHaveLength(count);
    expect(imp.rows.map((r: any) => r.source_number)).toEqual(
      Array.from({ length: 3 }, () =>
        Array.from({ length: count / 3 }, (_, i) => i + 1),
      ).flat(),
    );
    const types =
      file === "pdf.pdf"
        ? [
            "multiple",
            "multiple",
            "multiple",
            ...Array(4).fill("single"),
            ...Array(6).fill("numerical"),
            ...Array(4).fill("single"),
          ]
        : [
            ...Array(4).fill("single"),
            ...Array(3).fill("multiple"),
            ...Array(6).fill("numerical"),
            ...Array(count / 3 - 13).fill("single"),
          ];
    expect(imp.rows.map((r: any) => r.question.type)).toEqual(
      Array.from({ length: 3 }, () => types).flat(),
    );
    const marks =
      file === "2025_1_English.pdf"
        ? [...Array(4).fill("3"), ...Array(12).fill("4")]
        : file === "pdf.pdf"
          ? [
              ...Array(3).fill("4"),
              ...Array(4).fill("3"),
              ...Array(6).fill("4"),
              ...Array(4).fill("3"),
            ]
          : [
              ...Array(4).fill("3"),
              ...Array(9).fill("4"),
              ...Array(4).fill("3"),
            ];
    const penalties =
      file === "pdf.pdf"
        ? [
            ...Array(3).fill("2"),
            ...Array(4).fill("1"),
            ...Array(6).fill("0"),
            ...Array(4).fill("1"),
          ]
        : [
            ...Array(4).fill("1"),
            ...Array(3).fill("2"),
            ...Array(6).fill("0"),
            ...Array(count / 3 - 13).fill("1"),
          ];
    for (const [index, row] of imp.rows.entries()) {
      expect(row.question.marks).toBe(marks[index % marks.length]);
      expect(row.question.penalty).toBe(penalties[index % penalties.length]);
      expect(Number(row.question.marks)).toBeGreaterThan(0);
      expect(Number(row.question.penalty)).toBeGreaterThanOrEqual(0);
      expect(Number(row.question.penalty)).toBeLessThanOrEqual(
        Number(row.question.marks),
      );
      if (file !== "2025_1_English.pdf") {
        expect(row.question.key ?? []).toHaveLength(0);
        expect(row.question.numeric?.values ?? []).toHaveLength(0);
      }
    }
    if (file === "2025_1_English.pdf") {
      expect(imp.rows.filter((r: any) => r.error)).toHaveLength(2);
      expect(imp.rows[0].question.key).toEqual(["C"]);
      expect(imp.rows[4].question.key).toEqual(["A", "C"]);
      expect(imp.rows[7].question.numeric.values).toEqual(["105"]);
      expect(imp.rows[8].question.numeric).toMatchObject({
        min: "1.15",
        max: "1.25",
      });
      expect(imp.rows[26].question.numeric.values).toEqual(["0.5", "0.75"]);
    }
    expect(imp.raw_text.length).toBeGreaterThan(1000);
    await page.reload();
    await expect(page.locator(".draft-question")).toHaveCount(count);
    await page.screenshot({
      path: `${artifacts}/upload-${file}.png`,
      fullPage: true,
    });
  });
}

test("owner OMR images match independent visual transcriptions without inventing student grades", async ({
  page,
}) => {
  test.setTimeout(120000);
  await signInStaff(page);
  const base = `/api/v1/institutions/${fixture.org}/education/`;
  const me = await (await page.request.get("/api/v1/me")).json();
  const refs = JSON.parse(
    readFileSync(`${artifacts}/omr-references.json`, "utf8"),
  );
  for (const ref of refs) {
    const response = await page.request.post(base + "grid-preview", {
      headers: { "X-CSRF-Token": me.csrf, Origin: "http://localhost:3000" },
      multipart: {
        course_id: fixture.course,
        calibration: JSON.stringify(ref.calibration),
        file: {
          name: ref.file,
          mimeType: "image/jpeg",
          buffer: readFileSync(resolve(source, ref.file)),
        },
      },
    });
    expect(response.ok()).toBe(true);
    const reading = await response.json();
    const mismatches = reading.rows.filter(
      (r: any, i: number) =>
        JSON.stringify(r.selection) !==
        JSON.stringify(ref.expected[i] ? [ref.expected[i]] : []),
    );
    record(`omr-${ref.file}`, {
      calibration: ref.calibration,
      expected: ref.expected,
      reading,
      mismatches,
    });
    expect(reading.rows).toHaveLength(ref.expected.length);
    expect(mismatches).toHaveLength(0);
    console.log(
      `${ref.file}: ${reading.rows.length} independently checked positions; ${reading.rows.filter((r: any) => r.uncertain).length} review flags; no student identity or real grade inferred.`,
    );
  }
});
