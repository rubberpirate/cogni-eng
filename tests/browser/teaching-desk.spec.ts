import { signInStaff } from "./staff-sign-in";
import { expect, test, type Page } from "@playwright/test";
import { readFileSync } from "node:fs";
import { navigate } from "./navigation";
const creds = readFileSync(
  process.env.E2E_ACCESS_FILE ?? ".local/browser-access.txt",
  "utf8",
);
const fixture = JSON.parse(
  readFileSync(".local/academic-browser-access.json", "utf8"),
);
async function signIn(page: Page) {
  await signInStaff(page);
  await expect(
    page.getByRole("heading", { name: "Draft assessment", exact: true }),
  ).toBeVisible();
  await navigate(page, "Home");
}
async function createTopic(page: Page) {
  const me = await (await page.request.get("/api/v1/me")).json();
  const courses = await (
    await page.request.get(
      `/api/v1/institutions/${fixture.org}/education/courses`,
    )
  ).json();
  const course = courses.find((c: { id: string }) => c.id === fixture.course);
  const response = await page.request.post(
    `/api/v1/institutions/${fixture.org}/education/topics`,
    {
      headers: { "X-CSRF-Token": me.csrf, Origin: "http://localhost:3000" },
      data: {
        id: "",
        campus_id: course.campus_id,
        course_id: course.id,
        subject: "Mathematics",
        name: `Teaching desk ${Date.now()}`,
        outcome: "Identify simple numbers and shapes.",
        language: "en",
      },
    },
  );
  expect(response.ok()).toBe(true);
  return response.json();
}
test("manual entry and range/list marking persist through real Go APIs", async ({
  page,
}) => {
  await signIn(page);
  await expect(
    page.getByRole("button", { name: /02 Enter manually/ }),
  ).toBeVisible();
  const topic = await createTopic(page);
  await page.reload();
  await navigate(page, "Enter questions manually");
  await page.getByLabel("Topic", { exact: true }).selectOption(topic.id);
  await page
    .getByRole("button", { name: "Add a question", exact: true })
    .click();
  await page
    .getByLabel("Question wording", { exact: true })
    .fill("How many sides does a square have?");
  for (const [id, text] of [
    ["A", "Three"],
    ["B", "Four"],
    ["C", "Five"],
    ["D", "Six"],
  ])
    await page.getByLabel(`Option ${id}`, { exact: true }).fill(text);
  await page.getByLabel("Correct answer B", { exact: true }).check();
  await page
    .getByLabel("Worked solution / marking guidance", { exact: true })
    .fill("A square has four equal sides.");
  await page
    .getByRole("button", { name: "Save question draft", exact: true })
    .click();
  await expect(
    page.getByRole("button", { name: "Add a question", exact: true }),
  ).toBeVisible();
  await navigate(page, "Upload paper");
  await page.getByLabel("Question file", { exact: true }).setInputFiles({
    name: `range-${Date.now()}.csv`,
    mimeType: "text/csv",
    buffer: Buffer.from(
      "type,question,A,B,C,D,answer,marks,penalty,solution\nsingle,How many sides does a triangle have?,Three,Four,Five,Six,A,1,0,A triangle has three sides.\nsingle,What is two plus two?,Two,Three,Four,Five,C,1,0,Two plus two equals four.\nsingle,What follows five?,Three,Four,Five,Six,D,1,0,Six follows five.\n",
    ),
  });
  await page
    .getByRole("button", { name: "Upload for preview", exact: true })
    .click();
  await expect(page.locator(".import-question")).toHaveCount(3);
  await page.getByLabel("Question numbers", { exact: true }).fill("1-2,3");
  await page.getByLabel("Marks for correct answer", { exact: true }).fill("4");
  await page
    .getByLabel("Deduction for wrong answer", { exact: true })
    .fill("0.5");
  await page.getByRole("button", { name: "Apply marks", exact: true }).click();
  await expect(
    page.getByText(
      "Marking rules saved. Review the questions before saving to the bank.",
      { exact: true },
    ),
  ).toBeVisible();
  await page.getByLabel("Question numbers", { exact: true }).fill("2");
  await page.getByLabel("Marks for correct answer", { exact: true }).fill("2");
  await page.getByRole("button", { name: "Apply marks", exact: true }).click();
  await expect(
    page.getByLabel("Question 2 marks", { exact: true }),
  ).toHaveValue("2");
  await page.reload();
  await expect(
    page.getByLabel("Question 1 marks", { exact: true }),
  ).toHaveValue("4");
  await expect(
    page.getByLabel("Question 2 marks", { exact: true }),
  ).toHaveValue("2");
  await expect(
    page.getByLabel("Question 3 penalty", { exact: true }),
  ).toHaveValue("0.5");
  await page
    .getByRole("button", {
      name: "Save reviewed questions to bank",
      exact: true,
    })
    .click();
  await expect(
    page.getByRole("heading", {
      name: "Import review · committed",
      exact: true,
    }),
  ).toBeVisible();
  const questions = await (
    await page.request.get(
      `/api/v1/institutions/${fixture.org}/education/questions?topic=${topic.id}`,
    )
  ).json();
  expect(questions).toHaveLength(4);
  expect(
    questions.find((q: { stem: string }) => q.stem.includes("square")).key,
  ).toEqual(["B"]);
});
test("empty extraction gives recovery rather than a ready success state", async ({
  page,
}) => {
  await signIn(page);
  const topics = await (
    await page.request.get(
      `/api/v1/institutions/${fixture.org}/education/topics`,
    )
  ).json();
  const empty = {
    id: "empty-desk-fixture",
    topic_id: topics[0].id,
    state: "preview",
    rows: [],
    raw_text: "Unrecognized source text",
    version: 1,
    created_at: new Date().toISOString(),
  };
  await page.route("**/education/imports/empty-desk-fixture", (route) =>
    route.fulfill({ json: empty }),
  );
  await navigate(page, "Upload paper");
  await page.evaluate((hash) => {
    window.location.hash = hash;
  }, `education/import?import=${empty.id}&topic=${empty.topic_id}`);
  await page.reload();
  await expect(
    page.getByRole("region", {
      name: "No questions detected — review the extracted text",
      exact: true,
    }),
  ).toBeVisible();
  await expect(
    page.getByText("0 questions ready for review", { exact: true }),
  ).toHaveCount(0);
  await expect(
    page.getByRole("button", {
      name: "Save reviewed questions to bank",
      exact: true,
    }),
  ).toBeDisabled();
  await page.screenshot({
    path: ".local/screenshots/desk-empty-extraction.png",
    fullPage: true,
  });
});

test("batch OMR accepts separate files and retains matched identities for review", async ({
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
  const response = await page.request.get((await link.getAttribute("href"))!);
  expect(response.ok()).toBe(true);
  const { execFileSync } = await import("node:child_process");
  const archive = await response.body();
  const image = (name: string) =>
    execFileSync(
      "python3",
      [
        "-c",
        "import sys,zipfile,io;sys.stdout.buffer.write(zipfile.ZipFile(io.BytesIO(sys.stdin.buffer.read())).read(sys.argv[1]))",
        name,
      ],
      { input: archive, maxBuffer: 8 << 20 },
    );
  await page
    .getByRole("button", { name: "Go to Scan & check", exact: true })
    .click();
  await page.getByLabel("Batch upload", { exact: true }).check();
  await expect(
    page.getByLabel("Scanned PDF or page image", { exact: true }),
  ).toHaveAttribute("multiple", "");
  await page
    .getByLabel("Scanned PDF or page image", { exact: true })
    .setInputFiles([
      {
        name: "correct.png",
        mimeType: "image/png",
        buffer: image("omr-correct.png"),
      },
      {
        name: "review.png",
        mimeType: "image/png",
        buffer: image("omr-needs-review.png"),
      },
      {
        name: "correct-again.png",
        mimeType: "image/png",
        buffer: image("omr-correct.png"),
      },
    ]);
  await page
    .getByRole("button", { name: "Upload and read batch", exact: true })
    .click();
  await expect(page.locator(".batch-status > div")).toHaveCount(2);
  await expect(page.locator(".scan-review")).toHaveCount(2, { timeout: 25000 });
  const learner = await page
    .getByLabel("Assigned learner", { exact: true })
    .inputValue();
  expect(learner).not.toBe("");
  for (const select of await page
    .getByLabel("Confirmed assigned learner", { exact: true })
    .all())
    await expect(select).toHaveValue(learner);
  await expect(
    page.getByText("Multiple circles in a single-selection question.", {
      exact: false,
    }),
  ).toBeVisible();
  await page.screenshot({
    path: ".local/screenshots/desk-batch-review.png",
    fullPage: true,
  });
});
