import { test, expect } from "@playwright/test";
import { readFileSync } from "node:fs";
import { execFileSync } from "node:child_process";
import { signInStaff } from "./staff-sign-in";
import { navigate } from "./navigation";
import { originalOMRAssessment, savedOMRTemplate } from "./omr-fixtures";
test.skip(
  process.env.RUN_OMRCHECKER_TESTS !== "1",
  "Actual installed native engine and local OCR required; no recognition fixture.",
);
test("premade templates handle one image, a multipage PDF and a partially rejected batch", async ({
  page,
}) => {
  test.setTimeout(120000);
  await signInStaff(page);
  const { paper, base } = await originalOMRAssessment(page);
  const template = await savedOMRTemplate(page, paper, base);
  await page.reload();
  await expect(
    page.getByRole("group", { name: "Workspace tools", exact: true }),
  ).toBeEnabled();
  await navigate(page, "Upload OMR answer sheets");
  await page
    .getByLabel("1. Assessment", { exact: true })
    .selectOption(paper.id);
  await page
    .getByLabel("2. OMR template", { exact: true })
    .selectOption(template.id);
  await expect(
    page.getByLabel("Template sheet image", { exact: true }),
  ).not.toBeVisible();
  const image = readFileSync(
    "apps/web/public/samples/omrchecker/filled-sheet.png",
  );
  await page.getByLabel("3. Answer sheets", { exact: true }).setInputFiles([
    { name: "one-student.png", mimeType: "image/png", buffer: image },
    {
      name: "unreadable.png",
      mimeType: "image/png",
      buffer: Buffer.from("This is not an image."),
    },
  ]);
  await page
    .getByRole("button", { name: "Check answer sheets", exact: true })
    .click();
  await expect(
    page.getByRole("button", { name: "Retry failed uploads", exact: true }),
  ).toBeVisible();
  await expect(
    page.getByRole("alert").filter({ hasText: "unreadable.png" }),
  ).toContainText("Use PDF, PNG or JPEG");
  let checks: any[] = [];
  await expect
    .poll(
      async () => {
        checks = await (
          await page.request.get(base + "omr-checks?paper=" + paper.id)
        ).json();
        return checks.length === 1 && checks[0].state === "completed";
      },
      { timeout: 45000 },
    )
    .toBe(true);
  expect(checks[0].sheets[0]).toMatchObject({
    registration_number: "123456",
    score: "76",
    maximum: "80",
    state: "needs_review",
  });
  const pdf = execFileSync(
    "omr/.venv/bin/python",
    [
      "-c",
      "from PIL import Image;import sys,io;im=Image.open(io.BytesIO(sys.stdin.buffer.read())).convert('RGB');b=io.BytesIO();im.save(b,format='PDF',resolution=144,save_all=True,append_images=[im]);sys.stdout.buffer.write(b.getvalue())",
    ],
    { input: image, maxBuffer: 8 << 20 },
  );
  await page
    .getByLabel("3. Answer sheets", { exact: true })
    .setInputFiles({
      name: "two-students.pdf",
      mimeType: "application/pdf",
      buffer: pdf,
    });
  await page
    .getByRole("button", { name: "Check answer sheets", exact: true })
    .click();
  await expect
    .poll(
      async () => {
        checks = await (
          await page.request.get(base + "omr-checks?paper=" + paper.id)
        ).json();
        return (
          checks.length === 2 && checks.every((c) => c.state === "completed")
        );
      },
      { timeout: 45000 },
    )
    .toBe(true);
  const sheets = checks.flatMap((c) => c.sheets);
  expect(sheets).toHaveLength(3);
  expect(
    sheets.every((s) => s.score === "76" && s.state === "needs_review"),
  ).toBe(true);
  await page.reload();
  await expect(page.getByLabel("1. Assessment", { exact: true })).toHaveValue(
    paper.id,
  );
  await expect(
    page
      .getByRole("region", {
        name: "OMR grades by registration number",
        exact: true,
      })
      .getByRole("row"),
  ).toHaveCount(4);
  await page
    .locator(".omr-saved-results")
    .screenshot({ path: ".local/screenshots/omr-pdf-batch-results.png" });
});
