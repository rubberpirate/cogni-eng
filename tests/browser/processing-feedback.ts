import { expect, type Page } from "@playwright/test";
import { navigate } from "./navigation";

export async function checkProcessingFeedback(
  page: Page,
  fixture: { org: string; course: string },
) {
  await page.setViewportSize({ width: 1280, height: 900 });
  const topics = await (
    await page.request.get(
      `/api/v1/institutions/${fixture.org}/education/topics`,
    )
  ).json();
  const topic = topics.find(
    (t: { course_id: string }) => t.course_id === fixture.course,
  );
  if (!topic) throw new Error("An authored course topic is required");
  let state = "queued";
  const original = {
    id: "processing-ui",
    topic_id: topic.id,
    asset_id: "",
    state: "extracting",
    rows: [],
    raw_text: "",
    version: 1,
    created_at: new Date().toISOString(),
  };
  const job = () => ({
    id: "processing-job-ui",
    resource_id: original.id,
    kind: "extract_document",
    state,
    progress: state === "running" ? 35 : 0,
    input: {},
    output: {},
    error:
      state === "failed"
        ? "Model service unavailable. Saved source retained."
        : "",
    attempts: 1,
    provider: "",
    model: "",
    actor_id: "fixture",
    course_id: fixture.course,
    campus_id: "fixture",
    created_at: original.created_at,
  });
  let releaseUpload!: () => void;
  const uploadGate = new Promise<void>((resolve) => (releaseUpload = resolve));
  await page.route("**/education/imports/stage/upload", async (route) => {
    await uploadGate;
    await route.fulfill({ json: original });
  });
  await page.route(/\/education\/jobs(?:\?.*)?$/, (route) =>
    route.fulfill({ json: [job()] }),
  );
  await page.route("**/education/imports/processing-ui", (route) =>
    route.fulfill({
      json:
        state === "completed"
          ? {
              ...original,
              state: "preview",
              rows: [
                {
                  index: 1,
                  source_number: 1,
                  error: "supply the approved selection key",
                  question: {
                    topic_id: topic.id,
                    type: "single",
                    stem: "How many sides does a triangle have?",
                    options: [
                      { id: "A", text: "Three" },
                      { id: "B", text: "Four" },
                    ],
                    key: [],
                    marks: "1",
                    penalty: "0",
                    solution: "Three sides.",
                    language: "en",
                    provenance: "Original UI fixture",
                    license: "CC0",
                    difficulty: "standard",
                  },
                },
              ],
            }
          : original,
    }),
  );
  await page.route(
    "**/education/jobs/processing-job-ui/retry",
    async (route) => {
      state = "running";
      await route.fulfill({ json: {} });
    },
  );
  await navigate(page, "Upload paper");
  await expect(
    page.locator('#workspace-navigation [aria-current="page"]'),
  ).toHaveText("Draft assessment");
  const setup = page.getByText("Upload another paper or change topic", {
    exact: true,
  });
  if (await setup.isVisible()) await setup.click();
  await page.getByLabel("Topic", { exact: true }).selectOption(topic.id);
  await page.getByLabel("Question file", { exact: true }).setInputFiles({
    name: "original-progress.pdf",
    mimeType: "application/pdf",
    buffer: Buffer.from("Explicit UI fixture; not a model response"),
  });
  await page
    .getByRole("button", { name: "Upload for preview", exact: true })
    .click();
  await expect(
    page.getByRole("region", {
      name: "Uploading your question paper",
      exact: true,
    }),
  ).toBeVisible();
  await expect(
    page.getByRole("progressbar", {
      name: "Uploading your question paper progress",
      exact: true,
    }),
  ).not.toHaveAttribute("value");
  releaseUpload();
  await expect(
    page.getByRole("region", {
      name: "Paper saved — waiting to extract",
      exact: true,
    }),
  ).toBeVisible();
  await expect(page.getByLabel("Question file", { exact: true })).toBeHidden();
  await expect(
    page.getByText("Import preview saved; no questions have been approved.", {
      exact: true,
    }),
  ).toHaveCount(0);
  state = "running";
  const progress = page.getByRole("progressbar", {
    name: "Extracting questions and options progress",
    exact: true,
  });
  await expect(progress).toHaveAttribute("value", "35", { timeout: 7000 });
  await page.evaluate(() => window.scrollTo(0, 0));
  await page.screenshot({
    path: ".local/screenshots/processing-paper-desktop.png",
    fullPage: true,
  });
  await page.setViewportSize({ width: 390, height: 844 });
  await page.emulateMedia({ reducedMotion: "reduce" });
  expect(
    await page
      .locator(".task-spinner")
      .evaluate((el) => getComputedStyle(el).animationName),
  ).toBe("none");
  expect(
    await page.evaluate(
      () => document.documentElement.scrollWidth <= innerWidth,
    ),
  ).toBe(true);
  await page.screenshot({
    path: ".local/screenshots/processing-paper-mobile.png",
    fullPage: true,
  });
  await page.evaluate(() => {
    document.documentElement.dataset.theme = "dark";
    window.scrollTo(0, 0);
  });
  await page.screenshot({
    path: ".local/screenshots/processing-paper-dark.png",
    fullPage: true,
  });
  state = "failed";
  await expect(
    page.getByRole("button", { name: "Retry paper extraction", exact: true }),
  ).toBeVisible({ timeout: 7000 });
  await page
    .getByRole("button", { name: "Retry paper extraction", exact: true })
    .click();
  await expect(progress).toBeVisible();
  state = "completed";
  await expect(
    page.getByRole("heading", {
      name: "1 question ready for review",
      exact: true,
    }),
  ).toBeVisible({ timeout: 7000 });
  await expect(
    page.getByLabel("Question 1 option A", { exact: true }),
  ).toHaveValue("Three");
  await expect(page.getByRole("progressbar")).toHaveCount(0);
  await navigate(page, "Scan & check");
  await expect(
    page.locator('#workspace-navigation [aria-current="page"]'),
  ).toHaveText("Upload OMR answer sheets");
  await expect(
    page.getByRole("heading", { name: "Upload questions", exact: true }),
  ).toHaveCount(0);
}
