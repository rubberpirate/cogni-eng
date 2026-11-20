import { expect, type Page } from "@playwright/test";
import { readFileSync } from "node:fs";
export const omrFixture = JSON.parse(
  readFileSync(".local/academic-browser-access.json", "utf8"),
);
export const nativeExpected = JSON.parse(
  readFileSync(
    "apps/web/public/samples/omrchecker/expected-answers.json",
    "utf8",
  ),
).responses as Record<string, string>;
export async function originalOMRAssessment(page: Page) {
  const base = `/api/v1/institutions/${omrFixture.org}/education/`;
  const me = await (await page.request.get("/api/v1/me")).json();
  async function post(path: string, data: unknown) {
    const r = await page.request.post(base + path, {
      headers: { "X-CSRF-Token": me.csrf, Origin: "http://localhost:3000" },
      data,
    });
    const b = await r.json();
    expect(r.status(), JSON.stringify(b.message ?? "")).toBe(200);
    return b;
  }
  const courses = await (await page.request.get(base + "courses")).json();
  const course = courses.find((c: any) => c.id === omrFixture.course);
  expect(course).toBeTruthy();
  const stamp = crypto.randomUUID().slice(0, 8);
  const topic = await post("topics", {
    campus_id: course.campus_id,
    course_id: course.id,
    subject: "Recognition practice",
    name: "Original OMR recognition " + stamp,
    outcome: "Read documented original marked responses",
    language: "en",
  });
  const questions = Array.from({ length: 20 }, (_, i) => {
    const key = nativeExpected[`q${i + 1}`] || "B";
    return {
      topic_id: topic.id,
      type: i === 11 ? "multiple" : "single",
      stem: `Original recognition exercise ${stamp}-${i + 1}: select ${key} as the documented practice response.`,
      options: ["A", "B", "C", "D"].map((id) => ({ id, text: `Option ${id}` })),
      key: key.split(""),
      rubric: [],
      solution:
        "This original synthetic exercise checks the documented marked response.",
      marks: "4",
      penalty: "1",
      difficulty: "standard",
      language: "en",
      provenance: "Original CC0 recognition exercise",
      license: "CC0-1.0",
    };
  });
  const draft = await post("drafts", {
    campus_id: course.campus_id,
    course_id: course.id,
    version: 0,
    spec: {
      title: "OMR checking " + stamp,
      method: "manual",
      topic_id: topic.id,
      exam: "Original recognition practice",
      starts_at: new Date().toISOString(),
      duration_minutes: 30,
      instructions: "Original synthetic known-key recognition exercise",
      learners: [],
      questions,
      import_id: "",
      job_ids: [],
      applied_jobs: [],
      source_applied: false,
    },
  });
  const published = await post(`drafts/${draft.id}/publish`, {
    version: draft.version,
    reviewed: true,
  });
  const paper = await (
    await page.request.get(base + "papers/" + published.paper_id)
  ).json();
  const assignments = await (
    await page.request.get(base + "assignments?paper=" + paper.id)
  ).json();
  expect(assignments).toHaveLength(0);
  return { paper, base, post };
}

export async function savedOMRTemplate(page: Page, paper: any, base: string) {
  const { nativeFiles } = await import("../../clients/omr-template");
  const block = (first: number, x: number) => ({
    first,
    count: 10,
    options: 4,
    size: 0.38,
    bounds: {
      x: (x - 30) / 1200,
      y: 720 / 1600,
      width: 270 / 1200,
      height: 696 / 1600,
    },
    points: [750, 820].flatMap((y) =>
      [x, x + 70, x + 140, x + 210].map((x) => ({ x: x / 1200, y: y / 1600 })),
    ),
  });
  const layout = {
    name: "Premade practice " + paper.id.slice(0, 8),
    width: 1200,
    height: 1600,
    identity: {
      student_name: { x: 0.06, y: 0.1, width: 0.66, height: 0.028 },
      roll_number: { x: 0.06, y: 0.13, width: 0.34, height: 0.027 },
    },
    blocks: [block(1, 180), block(11, 740)],
  };
  const files = nativeFiles(layout);
  const me = await (await page.request.get("/api/v1/me")).json();
  const response = await page.request.post(base + "omr-templates", {
    headers: { "X-CSRF-Token": me.csrf, Origin: "http://localhost:3000" },
    multipart: {
      paper_id: paper.id,
      definition: JSON.stringify({
        name: layout.name,
        ...files,
        identity: Object.entries(layout.identity).map(([kind, rect]) => ({
          kind,
          rect,
        })),
        layout,
      }),
      file: {
        name: "reference.png",
        mimeType: "image/png",
        buffer: readFileSync(
          "apps/web/public/samples/omrchecker/filled-sheet.png",
        ),
      },
    },
  });
  const value = await response.json();
  expect(response.status(), JSON.stringify(value.message ?? "")).toBe(200);
  return value;
}
