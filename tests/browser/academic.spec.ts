import { signInStaff } from "./staff-sign-in";
import { navigate } from "./navigation";
import { registeredScanAPI } from "./registered-scan-api";
import { test, expect } from "@playwright/test";
import { readFileSync } from "node:fs";
const creds = readFileSync(
  process.env.E2E_ACCESS_FILE ?? ".local/browser-access.txt",
  "utf8",
);
const ownerEmail = creds.match(/Email: (.+)/)![1],
  ownerPassword = creds.match(/Password: (.+)/)![1];
const fixture = JSON.parse(
  readFileSync(".local/academic-browser-access.json", "utf8"),
) as {
  org: string;
  course_name: string;
  learner: string;
  learner_email: string;
  learner_password: string;
};
const local = (at: number) =>
  new Date(at - new Date().getTimezoneOffset() * 60000)
    .toISOString()
    .slice(0, 16);
test("teacher imports and approves a key, publishes and assigns papers; scans and online answers create real results", async ({
  page,
  browser,
}) => {
  test.setTimeout(100000);
  await signInStaff(page);
  await navigate(page, "Question bank");
  await page.getByText("Add a topic", { exact: true }).click();
  const topicForm = page
    .getByRole("button", { name: "Create topic", exact: true })
    .locator("..");
  await topicForm
    .getByLabel("Course", { exact: true })
    .selectOption({ label: fixture.course_name + " / 2026–27" });
  await topicForm.getByLabel("Subject", { exact: true }).fill("Mathematics");
  await topicForm
    .getByLabel("Topic name", { exact: true })
    .fill("Seed arithmetic " + Date.now());
  await topicForm
    .getByLabel("Learning outcome", { exact: true })
    .fill("Add whole numbers and explain the sum.");
  await topicForm
    .getByRole("button", { name: "Create topic", exact: true })
    .click();
  await expect(
    page.getByRole("button", { name: "Author question", exact: true }),
  ).toBeEnabled();
  await page
    .getByRole("button", { name: "Upload questions", exact: true })
    .click();
  await page.getByLabel("Question file", { exact: true }).setInputFiles({
    name: "original.csv",
    mimeType: "text/csv",
    buffer: Buffer.from(
      "type,question,A,B,C,D,answer,marks,penalty,solution\nsingle,Seven seeds plus five seeds makes how many?,12,13,11,2,A,2,0.5,Seven plus five is twelve.\n",
    ),
  });
  await page
    .getByRole("button", { name: "Upload for preview", exact: true })
    .click();
  await expect(
    page.getByRole("heading", { name: "Import review · preview" }),
  ).toBeVisible();
  await page
    .getByRole("button", {
      name: "Save reviewed questions to bank",
      exact: true,
    })
    .click();
  await page
    .getByRole("button", {
      name: "Next: review imported questions",
      exact: true,
    })
    .click();
  await page.getByRole("button", { name: /^Approve Q-/ }).click();
  await page.getByLabel("Select for paper", { exact: true }).check();
  await page
    .getByRole("button", { name: "Compose with 1 selected questions" })
    .click();
  async function compose(title: string) {
    const form = page
      .getByRole("button", { name: "Save paper draft", exact: true })
      .locator("..");
    await form.getByLabel("Assessment title", { exact: true }).fill(title);
    await form
      .getByLabel("Result release", { exact: true })
      .selectOption("immediate");
    await form
      .getByRole("button", { name: "Save paper draft", exact: true })
      .click();
    await page
      .getByRole("button", { name: "Publish reviewed paper", exact: true })
      .click();
    const assignment = page
      .getByRole("button", { name: "Assign paper", exact: true })
      .locator("..");
    await assignment.getByLabel(/Academic Browser Learner/).check();
    await assignment
      .getByLabel("Opens (your device timezone)", { exact: true })
      .fill(local(Date.now() - 120000));
    await assignment
      .getByLabel("Closes (your device timezone)", { exact: true })
      .fill(local(Date.now() + 86400000));
    await assignment
      .getByRole("button", { name: "Assign paper", exact: true })
      .click();
    await expect(
      page.getByText("Assigned to the selected learners.", { exact: true }),
    ).toBeVisible();
  }
  const scanTitle = "Scanned arithmetic " + Date.now();
  await compose(scanTitle);
  await navigate(page, "Scan & check");
  const base = `/api/v1/institutions/${fixture.org}/education/`;
  const paperID = await page
    .getByLabel("1. Assessment", { exact: true })
    .inputValue();
  const me = await (await page.request.get("/api/v1/me")).json();
  const layoutResponse = await page.request.post(base + "layouts", {
    headers: { "X-CSRF-Token": me.csrf, Origin: "http://localhost:3000" },
    data: { paper_id: paperID, columns: 1 },
  });
  expect(layoutResponse.status()).toBe(200);
  const layout = await layoutResponse.json();
  const assignments = await (
    await page.request.get(base + "assignments?paper=" + paperID)
  ).json();
  const sheetURL =
    base + `sheet-image/${layout.id}?assignment=${assignments[0].id}`;
  const filled = await page.evaluate(
    async ({ url, center }) => {
      const response = await fetch(url);
      const bitmap = await createImageBitmap(await response.blob());
      const canvas = document.createElement("canvas");
      canvas.width = bitmap.width;
      canvas.height = bitmap.height;
      const ctx = canvas.getContext("2d")!;
      ctx.drawImage(bitmap, 0, 0);
      ctx.fillStyle = "black";
      ctx.beginPath();
      ctx.arc(center.x, center.y, 7, 0, Math.PI * 2);
      ctx.fill();
      return canvas.toDataURL("image/png").split(",")[1];
    },
    { url: sheetURL!, center: layout.spec.rows[0].centers[0] },
  );
  const legacy = await registeredScanAPI(page, paperID, layout.id);
  await legacy.upload(
    "filled-sheet.png",
    "image/png",
    Buffer.from(filled, "base64"),
  );
  const [scan] = await legacy.sheets();
  const corrections = Object.fromEntries(
    Object.entries(scan.machine.rows).map(([id, row]: [string, any]) => [
      id,
      row.answer,
    ]),
  );
  const committed = await legacy.review(scan, assignments[0].id, corrections);
  const result = await (
    await page.request.get(base + "sittings/" + committed.sitting_id)
  ).json();
  expect(result.result.score).toBe("2");
  await navigate(page, "Assessments");
  const onlineTitle = "Online arithmetic " + Date.now();
  await compose(onlineTitle);
  const learner = await browser.newContext({
    viewport: { width: 390, height: 844 },
  });
  const lp = await learner.newPage();
  await lp.goto("/sign-in");
  await lp
    .getByLabel("Email address", { exact: true })
    .fill(fixture.learner_email);
  await lp
    .getByLabel("Password", { exact: true })
    .fill(fixture.learner_password);
  await lp.getByRole("button", { name: "Sign in", exact: true }).click();

  const row = lp
    .getByText(onlineTitle, { exact: true })
    .locator("..")
    .locator("..");
  await row
    .getByRole("button", { name: "Start assessment", exact: true })
    .click();
  await lp.getByLabel("A. 12", { exact: true }).focus();
  await lp.keyboard.press("Space");
  await expect(
    lp.getByText("All recorded answers are saved.", { exact: true }),
  ).toBeVisible();
  await lp
    .getByRole("button", { name: "Submit saved answers", exact: true })
    .click();
  await expect(lp.getByText("2 / 2", { exact: true })).toBeVisible();
  await expect(
    lp.getByText("Seven plus five is twelve.", { exact: true }),
  ).toBeVisible();
  await expect(
    lp.getByRole("heading", {
      name: "Ask about this released question",
      exact: true,
    }),
  ).toBeVisible();
  expect(
    await lp.evaluate(
      () => document.documentElement.scrollWidth <= window.innerWidth,
    ),
  ).toBe(true);
  await expect(
    lp.getByText("Topic observations and evidence limits", { exact: true }),
  ).toBeVisible();
  await lp.screenshot({
    path: ".local/screenshots/academic-learner-result.png",
    fullPage: true,
  });
  await learner.close();
});
