import { test, expect } from "@playwright/test";
import { readFileSync } from "node:fs";
import { navigate } from "./navigation";
const creds = readFileSync(
  process.env.E2E_ACCESS_FILE ?? ".local/browser-access.txt",
  "utf8",
);
const fixture = JSON.parse(
  readFileSync(".local/academic-browser-access.json", "utf8"),
) as { org: string; course: string };

test("a two-column question image becomes ordered editable options; faculty selection and a separate uploaded key persist", async ({
  page,
  browser,
}) => {
  test.setTimeout(80000);
  const source = await browser.newPage({
    viewport: { width: 1200, height: 800 },
    deviceScaleFactor: 1,
  });
  const questions = [
    [
      "1",
      "How many sides does a triangle have?",
      ["Two", "Three", "Four", "Five"],
    ],
    ["2", "What is four plus two?", ["Six", "Seven", "Eight", "Nine"]],
    [
      "3",
      "Which whole number comes after eight?",
      ["Seven", "Eight", "Nine", "Ten"],
    ],
    [
      "4",
      "How many minutes make one hour?",
      ["Twenty", "Thirty", "Forty", "Sixty"],
    ],
  ] as const;
  await source.setContent(
    `<html><style>body{font:22px Arial;margin:50px}main{display:grid;grid-template-columns:1fr 1fr;gap:60px}article{margin-bottom:48px}p{margin:14px 0}</style><h1>Original practice questions</h1><main>${[questions.slice(0, 2), questions.slice(2)].map((col) => `<section>${col.map(([n, stem, options]) => `<article><p>Q. ${n}. ${stem}</p>${options.map((option, i) => `<p>(${String.fromCharCode(97 + i)}) ${option}</p>`).join("")}</article>`).join("")}</section>`).join("")}</main></html>`,
  );
  const image = await source.screenshot({ fullPage: true });
  await source.setContent(
    '<html><body style="font:32px Arial;padding:60px"><h1>Answer key</h1><p>2. A</p><p>3. C</p><p>4. D</p></body></html>',
  );
  const keyImage = await source.screenshot({ fullPage: true });
  await source.close();
  await page.goto("/sign-in");
  await page
    .getByLabel("Email address", { exact: true })
    .fill(creds.match(/Email: (.+)/)![1]);
  await page
    .getByLabel("Password", { exact: true })
    .fill(creds.match(/Password: (.+)/)![1]);
  await page.getByRole("button", { name: "Sign in", exact: true }).click();
  await expect(
    page.getByRole("heading", { name: "Home", exact: true }),
  ).toBeVisible();
  const me = await (await page.request.get("/api/v1/me")).json();
  const courses = await (
    await page.request.get(
      `/api/v1/institutions/${fixture.org}/education/courses`,
    )
  ).json();
  const course = courses.find((c: { id: string }) => c.id === fixture.course);
  const topic = await (
    await page.request.post(
      `/api/v1/institutions/${fixture.org}/education/topics`,
      {
        headers: { "X-CSRF-Token": me.csrf, Origin: "http://localhost:3000" },
        data: {
          id: "",
          campus_id: course.campus_id,
          course_id: course.id,
          subject: "Mathematics",
          name: "Original image review " + Date.now(),
          outcome: "Identify basic shape and number facts.",
          language: "en",
        },
      },
    )
  ).json();
  await navigate(page, "Question bank");
  await page
    .getByRole("button", { name: "Upload questions", exact: true })
    .click();
  await page.getByLabel("Topic", { exact: true }).selectOption(topic.id);
  await page.getByLabel("Question file", { exact: true }).setInputFiles({
    name: "original-columns.png",
    mimeType: "image/png",
    buffer: image,
  });
  await page
    .getByRole("button", { name: "Upload for preview", exact: true })
    .click();
  await expect(page.locator(".import-question")).toHaveCount(4, {
    timeout: 25000,
  });
  const importID = await page
    .getByLabel("Saved imports", { exact: true })
    .inputValue();
  const extracted = await (
    await page.request.get(
      `/api/v1/institutions/${fixture.org}/education/imports/${importID}`,
    )
  ).json();
  expect(extracted.raw_text).toContain("Original practice questions");
  for (let i = 1; i <= 4; i++)
    await expect(
      page.getByLabel(`Question ${i} option D`, { exact: true }),
    ).toBeVisible();
  await page.setViewportSize({ width: 390, height: 844 });
  expect(
    await page.evaluate(
      () => document.documentElement.scrollWidth <= innerWidth,
    ),
  ).toBe(true);
  await page
    .getByLabel("Question 1 wording", { exact: true })
    .fill("How many straight sides does a triangle have?");
  await page
    .locator(".import-question")
    .first()
    .getByLabel("Correct B", { exact: true })
    .check();
  await page
    .getByRole("button", { name: "Save reviewed question 1", exact: true })
    .click();
  await expect(
    page.getByText("Reviewed question saved.", { exact: true }),
  ).toBeVisible();
  await page
    .getByLabel("Separate answer key file", { exact: true })
    .setInputFiles({
      name: "faculty-key.png",
      mimeType: "image/png",
      buffer: keyImage,
    });
  await page
    .getByRole("button", { name: "Read separate answer key", exact: true })
    .click();
  const apply = page.getByRole("button", {
    name: "I reviewed this key — apply to questions",
    exact: true,
  });
  await expect(apply).toBeVisible({ timeout: 15000 });
  await expect(
    page
      .locator(".import-question")
      .nth(1)
      .getByLabel("Correct A", { exact: true }),
  ).not.toBeChecked();
  await apply.click();
  await expect(
    page
      .locator(".import-question")
      .nth(1)
      .getByLabel("Correct A", { exact: true }),
  ).toBeChecked();
  await expect(
    page
      .locator(".import-question")
      .nth(2)
      .getByLabel("Correct C", { exact: true }),
  ).toBeChecked();
  await expect(
    page
      .locator(".import-question")
      .nth(3)
      .getByLabel("Correct D", { exact: true }),
  ).toBeChecked();
  await expect(
    page.getByLabel("Question 1 wording", { exact: true }),
  ).toHaveValue("How many straight sides does a triangle have?");
  await page
    .getByRole("button", { name: "Commit clean rows", exact: true })
    .click();
  await expect(
    page.getByRole("heading", {
      name: "Import review · committed",
      exact: true,
    }),
  ).toBeVisible();
  const bank = await (
    await page.request.get(
      `/api/v1/institutions/${fixture.org}/education/questions?topic=${topic.id}`,
    )
  ).json();
  expect(bank).toHaveLength(4);
  expect(
    bank.find(
      (q: { stem: string }) =>
        q.stem === "How many straight sides does a triangle have?",
    ).key,
  ).toEqual(["B"]);
});
