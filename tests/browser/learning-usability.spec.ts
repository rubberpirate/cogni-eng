import { signInStaff } from "./staff-sign-in";
import { test, expect } from "@playwright/test";
import { readFileSync } from "node:fs";
import { execFileSync } from "node:child_process";
import { navigate } from "./navigation";
import { registeredScanAPI } from "./registered-scan-api";
const creds = readFileSync(
  process.env.E2E_ACCESS_FILE ?? ".local/browser-access.txt",
  "utf8",
);
const fixture = JSON.parse(
  readFileSync(".local/academic-browser-access.json", "utf8"),
) as { org: string; course: string; learner: string; course_name: string };
async function signIn(page: import("@playwright/test").Page) {
  await signInStaff(page);
  await expect(
    page.getByRole("heading", { name: "Draft assessment", exact: true }),
  ).toBeVisible();
  await navigate(page, "Home");
}
test("sample kit downloads matching PDFs and its registered evidence remains compatible through the API", async ({
  page,
}) => {
  test.setTimeout(80000);
  await signIn(page);
  await expect(
    page
      .locator("#workspace-navigation")
      .getByRole("button", { name: "People", exact: true }),
  ).not.toBeVisible();
  await navigate(page, "Try sample files");
  await page
    .getByLabel("Sample course", { exact: true })
    .selectOption(fixture.course);
  await page
    .getByLabel("Test learner", { exact: true })
    .selectOption(fixture.learner);
  await page
    .getByRole("button", { name: "Create a sample practice test", exact: true })
    .click();
  const link = page.getByRole("link", {
    name: "Download complete test kit (.zip)",
    exact: true,
  });
  await expect(link).toBeVisible();
  const download = await page.request.get((await link.getAttribute("href"))!);
  expect(download.status()).toBe(200);
  expect(download.headers()["content-type"]).toContain("application/zip");
  const zip = await download.body();
  const pdf = execFileSync(
    "python3",
    [
      "-c",
      'import sys,zipfile,io;sys.stdout.buffer.write(zipfile.ZipFile(io.BytesIO(sys.stdin.buffer.read())).read("omr-mixed.pdf"))',
    ],
    { input: zip, maxBuffer: 8 << 20 },
  );
  expect(pdf.subarray(0, 4).toString()).toBe("%PDF");
  await page
    .getByRole("button", { name: "Go to Scan & check", exact: true })
    .click();
  const selected = await page
    .getByLabel("1. Assessment", { exact: true })
    .inputValue();
  const base = `/api/v1/institutions/${fixture.org}/education/`;
  const layouts = await (
    await page.request.get(base + "layouts?paper=" + selected)
  ).json();
  const assignments = await (
    await page.request.get(base + "assignments?paper=" + selected)
  ).json();
  const legacy = await registeredScanAPI(page, selected, layouts[0].id);
  await legacy.upload("omr-mixed.pdf", "application/pdf", pdf);
  const [sheet] = await legacy.sheets();
  const corrections = Object.fromEntries(
    Object.entries(sheet.machine.rows).map(([id, value]: [string, any]) => [
      id,
      value.answer,
    ]),
  );
  const committed = await legacy.review(sheet, assignments[0].id, corrections);
  const sitting = await (
    await page.request.get(base + "sittings/" + committed.sitting_id)
  ).json();
  expect(sitting.result.score).toBe("5");
  await navigate(page, "Results & review");
  await page
    .getByLabel("Assessment to review", { exact: true })
    .selectOption(selected);
  await page.getByText("Account-linked results", { exact: true }).click();
  const resultRow = page
    .locator(".account-linked-results .results-table tbody tr")
    .first();
  await expect(resultRow).toContainText("5 / 12");
  await expect(resultRow.locator(".correct-question-list")).toHaveText(
    "1, 3, 5",
  );
  await page
    .getByLabel("Search students", { exact: true })
    .fill("missing-student-name");
  await expect(
    page.getByText("No students match your search.", { exact: true }),
  ).toBeVisible();
  await page.getByLabel("Search students", { exact: true }).fill("");
  await page.setViewportSize({ width: 1440, height: 1000 });
  await page.screenshot({
    path: ".local/screenshots/desk-student-results.png",
    fullPage: true,
  });
});
test("generation failure remains understandable on the request screen and navigation supports browser back", async ({
  page,
}) => {
  await signIn(page);
  await navigate(page, "Create questions with AI");
  const topics = await (
    await page.request.get(
      `/api/v1/institutions/${fixture.org}/education/topics`,
    )
  ).json();
  if (!topics.length)
    throw new Error("The preceding academic fixture should supply a topic");
  await page
    .getByLabel("Topic to generate questions for", { exact: true })
    .selectOption(topics[0].id);
  const job = {
    id: "ui-failure-fixture",
    campus_id: "fixture",
    course_id: fixture.course,
    actor_id: "fixture",
    kind: "generate_questions",
    resource_id: topics[0].id,
    input: {},
    state: "queued",
    progress: 0,
    error: "",
    attempts: 0,
    output: {},
    provider: "",
    model: "",
    created_at: new Date().toISOString(),
  };
  await page.route("**/education/generate", async (route) => {
    const body = route.request().postDataJSON();
    expect(body.topic_id).toBe(topics[0].id);
    expect(body.count).toBe(3);
    await route.fulfill({ json: job });
  });
  await page.route("**/education/jobs/ui-failure-fixture", (route) =>
    route.fulfill({
      json: {
        ...job,
        state: "failed",
        error: "Gemini returned HTTP 503; saved work is retained",
      },
    }),
  );
  await page.getByLabel("How many questions?", { exact: true }).fill("3");
  await page
    .getByRole("button", { name: "Generate draft questions", exact: true })
    .click();
  await expect(
    page.getByRole("progressbar", {
      name: "Your request is queued progress",
      exact: true,
    }),
  ).not.toHaveAttribute("value");
  await expect(
    page.getByRole("heading", {
      name: "The AI service is temporarily unavailable",
      exact: true,
    }),
  ).toBeVisible({ timeout: 10000 });
  await expect(
    page.getByText(
      "Gemini could not complete this request. Wait a little, then retry the saved request below.",
      { exact: true },
    ),
  ).toBeVisible();
  await page.getByText("Technical details", { exact: true }).click();
  await expect(
    page.getByText("Gemini returned HTTP 503; saved work is retained", {
      exact: true,
    }),
  ).toBeVisible();
  await expect(
    page.getByRole("button", { name: "Retry this request", exact: true }),
  ).toBeVisible();
  await expect(page).toHaveURL(/education\/generate/);
  await navigate(page, "Question bank");
  await page.goBack();
  await expect(
    page.getByRole("heading", {
      name: "Create questions with AI",
      exact: true,
    }),
  ).toBeVisible();
  await expect(
    page.getByRole("heading", {
      name: "The AI service is temporarily unavailable",
      exact: true,
    }),
  ).toBeVisible();
  await page.setViewportSize({ width: 390, height: 844 });
  await page.getByRole("button", { name: "Menu", exact: true }).focus();
  await page.keyboard.press("Enter");
  await expect(page.locator("#workspace-navigation")).toBeVisible();
  expect(
    await page.evaluate(
      () => document.documentElement.scrollWidth <= innerWidth,
    ),
  ).toBe(true);
});

test("a teacher writes a numerical question through ordinary fields without editing JSON", async ({
  page,
}) => {
  await signIn(page);
  await navigate(page, "Question bank");
  const topics = await (
    await page.request.get(
      `/api/v1/institutions/${fixture.org}/education/topics`,
    )
  ).json();
  const topic = topics.find(
    (t: { course_id: string }) => t.course_id === fixture.course,
  );
  if (!topic)
    throw new Error("Expected a course topic from the authoring journey");
  await page.getByLabel("Topic", { exact: true }).selectOption(topic.id);
  await page
    .getByRole("button", { name: "Author question", exact: true })
    .click();
  await page
    .getByLabel("Question type", { exact: true })
    .selectOption("numerical");
  const wording =
    "Compute one tenth plus two tenths. Original test " +
    crypto.randomUUID().slice(0, 8);
  await page.getByLabel("Question wording", { exact: true }).fill(wording);
  await page.getByLabel("Accepted answer(s)", { exact: true }).fill("0.3");
  await page
    .getByLabel("Worked solution / marking guidance", { exact: true })
    .fill("One tenth plus two tenths is three tenths: 0.3.");
  const saved = page.waitForResponse(
    (r) =>
      r.url().endsWith("/education/questions") &&
      r.request().method() === "POST",
  );
  await page
    .getByRole("button", { name: "Save question draft", exact: true })
    .click();
  const response = await saved;
  expect(response.status()).toBe(200);
  const result = await response.json();
  expect(result.question.type).toBe("numerical");
  expect(result.question.numeric.values).toEqual(["0.3"]);
  await expect(
    page.locator("summary").filter({ hasText: wording }),
  ).toBeVisible();
});

test("registered evidence API refuses an ambiguous sample until its uncertain response is corrected", async ({
  page,
}) => {
  test.setTimeout(60000);
  await signIn(page);
  await navigate(page, "Try sample files");
  await page
    .getByLabel("Sample course", { exact: true })
    .selectOption(fixture.course);
  await page
    .getByLabel("Test learner", { exact: true })
    .selectOption(fixture.learner);
  await page
    .getByRole("button", { name: "Create a sample practice test", exact: true })
    .click();
  const link = page.getByRole("link", {
    name: "Download complete test kit (.zip)",
    exact: true,
  });
  await expect(link).toBeVisible();
  const zip = await (
    await page.request.get((await link.getAttribute("href"))!)
  ).body();
  const image = execFileSync(
    "python3",
    [
      "-c",
      'import sys,zipfile,io;sys.stdout.buffer.write(zipfile.ZipFile(io.BytesIO(sys.stdin.buffer.read())).read("omr-needs-review.png"))',
    ],
    { input: zip, maxBuffer: 8 << 20 },
  );
  await page
    .getByRole("button", { name: "Go to Scan & check", exact: true })
    .click();
  const selected = await page
    .getByLabel("1. Assessment", { exact: true })
    .inputValue();
  const base = `/api/v1/institutions/${fixture.org}/education/`;
  const layouts = await (
    await page.request.get(base + "layouts?paper=" + selected)
  ).json();
  const assignments = await (
    await page.request.get(base + "assignments?paper=" + selected)
  ).json();
  const legacy = await registeredScanAPI(page, selected, layouts[0].id);
  await legacy.upload("omr-needs-review.png", "image/png", image);
  const [sheet] = await legacy.sheets();
  await legacy.review(sheet, assignments[0].id, {}, 409);
  const paper = await (
    await page.request.get(base + "papers/" + selected)
  ).json();
  const corrections = Object.fromEntries(
    Object.entries(sheet.machine.rows).map(([id, value]: [string, any]) => [
      id,
      value.answer,
    ]),
  );
  corrections[paper.spec.sections[0].questions[0].id] = {
    ...sheet.machine.rows[paper.spec.sections[0].questions[0].id].answer,
    selection: ["B"],
    invalid: false,
  };
  const committed = await legacy.review(sheet, assignments[0].id, corrections);
  const sitting = await (
    await page.request.get(base + "sittings/" + committed.sitting_id)
  ).json();
  expect(sitting.result.score).toBe("12");
});
