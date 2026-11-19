import { signInStaff } from "./staff-sign-in";
import { test, expect, type Page } from "@playwright/test";
import { readFileSync } from "node:fs";
import { createServer, type Server } from "node:http";
const creds = readFileSync(
  process.env.E2E_ACCESS_FILE ?? ".local/browser-access.txt",
  "utf8",
);
const fixture = JSON.parse(
  readFileSync(".local/academic-browser-access.json", "utf8"),
);
let model: Server;
const question = (stem: string) => ({
  topic_id: "",
  type: "single",
  stem,
  options: [
    { id: "A", text: "Four" },
    { id: "B", text: "Five" },
    { id: "C", text: "Six" },
    { id: "D", text: "Seven" },
  ],
  key: ["A"],
  rubric: [],
  solution:
    "The original test fixture defines four as the answer; not a live model evaluation.",
  marks: "4",
  penalty: "1",
  difficulty: "standard",
  language: "en",
  provenance: "Explicit HTTP contract fixture",
  license: "CC0",
});
test.beforeAll(async () => {
  model = createServer((req, res) => {
    let data = "";
    req.on("data", (c) => (data += c));
    req.on("end", () => {
      const body = JSON.parse(data);
      const prompt = body.messages[1].content as string;
      let payload;
      if (prompt.includes("OCR text:")) {
        payload = {
          rows: [
            {
              source_number: 1,
              index: 1,
              question: question("How many sides does a square have?"),
            },
            {
              source_number: 2,
              index: 2,
              question: {
                ...question("How many sides does a triangle have?"),
                options: [
                  { id: "A", text: "Four" },
                  { id: "B", text: "Three" },
                  { id: "C", text: "Five" },
                  { id: "D", text: "Six" },
                ],
                key: [],
                solution: "A triangle has three sides.",
              },
            },
          ],
        };
      } else {
        const input = JSON.parse(prompt.split("Curriculum data: ")[1]);
        expect(input.exam).toBe("IIT-JEE");
        payload = {
          questions: Array.from({ length: input.count }, (_, i) =>
            question(
              `${input.topic.subject} / ${input.topic.name}: original fixture ${i + 1}?`,
            ),
          ),
        };
      }
      res.writeHead(200, { "Content-Type": "application/json" });
      res.end(
        JSON.stringify({
          done: true,
          done_reason: "stop",
          message: { content: JSON.stringify(payload) },
        }),
      );
    });
  });
  await new Promise<void>((resolve, reject) => {
    model.once("error", reject);
    model.listen(11434, "127.0.0.1", resolve);
  });
});
test.afterAll(async () => {
  await new Promise<void>((resolve) => model?.close(() => resolve()));
});
async function signIn(page: Page) {
  await signInStaff(page);
  await expect(
    page.getByRole("heading", { name: "Draft assessment", exact: true }),
  ).toBeVisible();
}
async function setup(page: Page, method: string) {
  await page
    .getByRole("button", { name: new RegExp(method) })
    .filter({ has: page.locator("strong") })
    .click();
  await page
    .getByLabel("Assessment name", { exact: true })
    .fill(`Wizard ${method} ${Date.now()}`);
  await page
    .getByLabel("Student batch", { exact: true })
    .selectOption(fixture.course);
  await page.getByLabel("Main subject", { exact: true }).fill("Mathematics");
  await page
    .getByLabel("Main topic", { exact: true })
    .fill("Shape facts " + Date.now());
  await page
    .getByRole("button", { name: "Continue to questions", exact: true })
    .click();
}
test("five primary actions and a persistent manual paper publish to the real student batch", async ({
  page,
}) => {
  test.setTimeout(80000);
  await signIn(page);
  const nav = page.locator("#workspace-navigation");
  await expect(nav.locator(":scope > button")).toHaveText([
    "Draft assessment",
    "Active assessments",
    "Upload OMR answer sheets",
    "Students",
    "Results",
  ]);
  await expect(page.locator(".draft-method-choice")).toHaveCount(3);
  await page.screenshot({
    path: ".local/screenshots/assessment-draft-start.png",
    fullPage: true,
  });
  await setup(page, "Enter manually");
  const title = await page.locator(".editor-toolbar strong").textContent();
  await page
    .getByRole("button", { name: "Enter question", exact: true })
    .click();
  await page
    .getByLabel("Question wording", { exact: true })
    .fill("How many sides does a square have?");
  for (const [id, option] of [
    ["A", "Three"],
    ["B", "Four"],
    ["C", "Five"],
    ["D", "Six"],
  ])
    await page.getByLabel(`Option ${id}`, { exact: true }).fill(option);
  await page.getByLabel("Correct answer B", { exact: true }).check();
  await page
    .getByLabel("Worked solution / marking guidance", { exact: true })
    .fill("A square has four sides.");
  await page
    .getByRole("button", { name: "Save question draft", exact: true })
    .click();
  await expect(page.locator(".draft-question")).toHaveCount(1);
  await page.getByLabel("All questions", { exact: true }).check();
  await page.getByLabel("Marks for correct answer", { exact: true }).fill("3");
  await page
    .getByLabel("Deduction for wrong answer", { exact: true })
    .fill("0.5");
  await page.getByRole("button", { name: "Apply marks", exact: true }).click();
  await expect(
    page.getByText("Marking scheme saved.", { exact: true }),
  ).toBeVisible();
  await page.reload();
  await expect(page.locator(".draft-question")).toContainText("+3 / −0.5");
  await page
    .getByLabel(
      "I reviewed every question, answer, solution and marking rule.",
      { exact: true },
    )
    .check();
  await page
    .getByRole("button", { name: "Publish & assign assessment", exact: true })
    .click();
  await expect(
    page.getByRole("heading", {
      name: "Assessment published and assigned",
      exact: true,
    }),
  ).toBeVisible();
  await page.getByRole("button", { name: "Done", exact: true }).click();
  await nav
    .getByRole("button", { name: "Active assessments", exact: true })
    .click();
  await expect(
    page
      .getByRole("button", { name: "Edit as new revision", exact: true })
      .first(),
  ).toBeVisible();
  await page.screenshot({
    path: ".local/screenshots/assessment-active.png",
    fullPage: true,
  });
  await page
    .locator(".academic-row")
    .filter({ hasText: title! })
    .getByRole("button", { name: "Edit as new revision", exact: true })
    .click();
  await expect(page.locator(".draft-question")).toHaveCount(1);
  await expect(page.locator(".draft-question")).toContainText("+3 / −0.5");
  await expect(page.locator(".editor-toolbar strong")).toContainText(
    "revision",
  );
  await nav
    .getByRole("button", { name: "Draft assessment", exact: true })
    .click();
  await expect(page.locator(".draft-method-choice")).toHaveCount(3);
  await nav.getByRole("button", { name: "Students", exact: true }).click();
  await page
    .getByLabel("Student batch", { exact: true })
    .selectOption(fixture.course);
  await expect(
    page.getByRole("region", { name: "Students in your batch" }),
  ).toContainText("Academic Browser Learner");
});
test("uploaded pixels pass through OCR and local-model contract into the same editor", async ({
  page,
  browser,
}) => {
  test.setTimeout(80000);
  await signIn(page);
  await setup(page, "Upload question paper");
  await expect(
    page.getByLabel("Question paper file", { exact: true }),
  ).toBeVisible();
  const imagePage = await browser.newPage();
  await imagePage.setContent(
    '<html><body style="font:32px sans-serif;padding:70px"><h1>Original assessment fixture</h1><p>1. How many sides does a square have?</p><p>A. Four</p><p>B. Five</p><p>C. Six</p><p>D. Seven</p><p>Answer: A</p><p>2. How many sides does a triangle have?</p><p>A. Four</p><p>B. Three</p><p>C. Five</p><p>D. Six</p></body></html>',
  );
  const image = await imagePage.screenshot({ fullPage: true });
  await imagePage.close();
  await page.getByLabel("Question paper file", { exact: true }).setInputFiles({
    name: "original-wizard.png",
    mimeType: "image/png",
    buffer: image,
  });
  await expect(page.locator(".draft-question")).toHaveCount(2, {
    timeout: 35000,
  });
  await expect(page.locator(".draft-question").first()).toContainText(
    "How many sides does a square have?",
  );
  await page
    .locator(".draft-question")
    .nth(1)
    .getByRole("button", { name: "Move up", exact: true })
    .click();
  await page.getByRole("button", { name: "Save changes", exact: true }).click();
  await expect(
    page.getByText("Paper changes saved.", { exact: true }),
  ).toBeVisible();
  await page
    .getByLabel("Separate answer key file", { exact: true })
    .setInputFiles({
      name: "reviewed-key.csv",
      mimeType: "text/csv",
      buffer: Buffer.from("1,A\n2,B\n"),
    });
  await page
    .getByRole("button", { name: "Read separate answer key", exact: true })
    .click();
  await page
    .getByRole("button", {
      name: "I reviewed this key — apply to questions",
      exact: true,
    })
    .click();
  await expect(
    page.locator(".draft-question").first().getByRole("radio").nth(1),
  ).toBeChecked();
  await expect(
    page.locator(".draft-question").nth(1).getByRole("radio").first(),
  ).toBeChecked();
  await page.getByRole("button", { name: "Save changes", exact: true }).click();
  await expect(
    page.getByText("Paper changes saved.", { exact: true }),
  ).toBeVisible();
  await page.reload();
  await expect(page.locator(".draft-question")).toHaveCount(2);
  await page.screenshot({
    path: ".local/screenshots/assessment-upload-editor.png",
    fullPage: true,
  });
  for (let i = 0; i < 2; i++)
    await page
      .locator(".draft-question")
      .last()
      .getByRole("button", { name: /Remove question/ })
      .click();
  await page.getByRole("button", { name: "Save changes", exact: true }).click();
  await expect(
    page.getByText("Paper changes saved.", { exact: true }),
  ).toBeVisible();
  await page.reload();
  await expect(page.locator(".draft-question")).toHaveCount(0);
  await expect(
    page.getByRole("button", { name: "Add manual question", exact: true }),
  ).toBeVisible();
  await page.waitForTimeout(3000);
  await expect(page.locator(".draft-question")).toHaveCount(0);
});
test("multiple local AI groups add questions and retain their counts after reload", async ({
  page,
}) => {
  test.setTimeout(80000);
  await signIn(page);
  await setup(page, "Generate with AI");
  await page.getByLabel("Group 1 subject", { exact: true }).fill("Mathematics");
  await page
    .getByLabel("Group 1 topic or topic range", { exact: true })
    .fill("Whole numbers");
  await page.getByLabel("Group 1 question count", { exact: true }).fill("3");
  await page
    .getByRole("button", { name: "Add another group", exact: true })
    .click();
  await page.getByLabel("Group 2 subject", { exact: true }).fill("Mathematics");
  await page
    .getByLabel("Group 2 topic or topic range", { exact: true })
    .fill("Plane geometry");
  await page.getByLabel("Group 2 question count", { exact: true }).fill("5");
  await page
    .getByLabel("Group 2 difficulty", { exact: true })
    .selectOption("challenging");
  await page
    .getByRole("button", { name: "Generate question groups", exact: true })
    .click();
  await expect(page.locator(".draft-question")).toHaveCount(8, {
    timeout: 35000,
  });
  await page.reload();
  await expect(page.locator(".draft-question")).toHaveCount(8);
  await page.setViewportSize({ width: 390, height: 844 });
  expect(
    await page.evaluate(
      () => document.documentElement.scrollWidth <= innerWidth,
    ),
  ).toBe(true);
  await page.screenshot({
    path: ".local/screenshots/assessment-ai-editor-mobile.png",
    fullPage: true,
  });
  for (let i = 0; i < 8; i++)
    await page
      .locator(".draft-question")
      .last()
      .getByRole("button", { name: /Remove question/ })
      .click();
  await page.getByRole("button", { name: "Save changes", exact: true }).click();
  await expect(
    page.getByText("Paper changes saved.", { exact: true }),
  ).toBeVisible();
  await page.reload();
  await expect(page.locator(".draft-question")).toHaveCount(0);
  await expect(
    page.getByRole("button", { name: "Add AI questions", exact: true }),
  ).toBeVisible();
  await page.waitForTimeout(3000);
  await expect(page.locator(".draft-question")).toHaveCount(0);
});
