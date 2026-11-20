import { test, expect, type Page } from "@playwright/test";
import { readFileSync } from "node:fs";
import { resolve } from "node:path";
import { execFileSync } from "node:child_process";
import { signInStaff } from "./staff-sign-in";
import { navigate } from "./navigation";
import { originalOMRAssessment } from "./omr-fixtures";
test.skip(
  process.env.RUN_OMRCHECKER_TESTS !== "1",
  "Actual installed OMRChecker and Tesseract required; no recognition fixture.",
);
const fixture = JSON.parse(
  readFileSync(".local/academic-browser-access.json", "utf8"),
);
const sheet = resolve("apps/web/public/samples/omrchecker/filled-sheet.png");
const surface = (page: Page) =>
  page.getByRole("group", { name: "Template calibration canvas", exact: true });
async function draw(page: Page, a: number, b: number, c: number, d: number) {
  await page.locator(".omr-sheet-stage").scrollIntoViewIfNeeded();
  const r = (await surface(page).boundingBox())!;
  await page.mouse.move(r.x + a * r.width, r.y + b * r.height);
  await page.mouse.down();
  await page.mouse.move(r.x + c * r.width, r.y + d * r.height, { steps: 10 });
  await page.mouse.up();
}
async function seed(page: Page, x: number) {
  for (const y of [750, 820])
    for (const dx of [0, 70, 140, 210]) {
      await page.locator(".omr-sheet-stage").scrollIntoViewIfNeeded();
      const r = (await surface(page).boundingBox())!;
      await page.mouse.click(
        r.x + ((x + dx) / 1200) * r.width,
        r.y + (y / 1600) * r.height,
      );
    }
  await expect(
    page.getByLabel("Questions in this box", { exact: true }),
  ).toHaveValue("10");
  await page.getByRole("button", { name: "Save frame", exact: true }).click();
}
test("assessment then saved template grades an individual and batch without student accounts", async ({
  page,
}) => {
  test.setTimeout(140000);
  await page.setViewportSize({ width: 1440, height: 1100 });
  await signInStaff(page);
  await expect(
    page.getByRole("heading", { name: "Draft assessment", exact: true }),
  ).toBeVisible();
  const { paper, base } = await originalOMRAssessment(page);
  await page.reload();
  await expect(
    page.getByRole("group", { name: "Workspace tools", exact: true }),
  ).toBeEnabled();
  await navigate(page, "Upload OMR answer sheets");
  await page
    .getByLabel("1. Assessment", { exact: true })
    .selectOption(paper.id);
  await expect(
    page.getByLabel("2. OMR template", { exact: true }),
  ).toBeVisible();
  await expect(
    page.getByLabel("3. Answer sheets", { exact: true }),
  ).toBeDisabled();
  await expect(
    page.getByRole("button", { name: "Use existing files", exact: true }),
  ).toHaveCount(0);
  await expect(
    page.getByText("Advanced pixel-reader tools", { exact: true }),
  ).toHaveCount(0);
  await page.getByLabel("1. Assessment", { exact: true }).selectOption("");
  await page.getByText("Create an OMR template", { exact: true }).click();
  await page
    .getByLabel("Template sheet image", { exact: true })
    .setInputFiles(sheet);
  await expect(surface(page)).toBeVisible();
  await page
    .getByRole("button", { name: "Expand editor", exact: true })
    .click();
  await expect(page.getByRole("dialog")).toBeVisible();
  await page
    .getByRole("button", { name: "Close expanded editor", exact: true })
    .click();
  await expect(page.getByRole("dialog")).not.toBeVisible();
  await expect(
    page.getByRole("button", { name: "Expand editor", exact: true }),
  ).toBeFocused();
  await expect(page.getByLabel("Layout zoom", { exact: true })).toHaveValue(
    "fit",
  );
  await expect(surface(page).locator("ellipse")).toHaveCount(0);
  await surface(page).focus();
  await surface(page).press("Enter");
  await surface(page).press("Shift+ArrowRight");
  await surface(page).press("Shift+ArrowDown");
  await surface(page).press("Enter");
  await expect(page.getByLabel("Field label", { exact: true })).toBeVisible();
  await page
    .getByRole("button", { name: "Discard frame", exact: true })
    .click();
  await draw(page, 0.06, 0.1, 0.72, 0.128);
  await page
    .getByLabel("Field label", { exact: true })
    .selectOption("student_name");
  await page.getByRole("button", { name: "Save frame", exact: true }).click();
  await draw(page, 0.06, 0.13, 0.4, 0.157);
  await page
    .getByLabel("Field label", { exact: true })
    .selectOption("roll_number");
  await page.getByRole("button", { name: "Save frame", exact: true }).click();
  await draw(page, 150 / 1200, 0.45, 420 / 1200, 0.885);
  await expect(surface(page).locator("ellipse")).toHaveCount(0);
  await page.getByLabel("Field label", { exact: true }).selectOption("answers");
  await expect(surface(page).locator("ellipse")).toHaveCount(0);
  await page
    .getByRole("button", { name: "Mark first two rows", exact: true })
    .click();
  await seed(page, 180);
  await draw(page, 710 / 1200, 0.45, 980 / 1200, 0.885);
  await page.getByLabel("Field label", { exact: true }).selectOption("answers");
  await page
    .getByRole("button", { name: "Mark first two rows", exact: true })
    .click();
  await seed(page, 740);
  await page
    .locator(".omr-designer")
    .screenshot({ path: ".local/screenshots/omr-visual-editor.png" });
  await page.setViewportSize({ width: 390, height: 844 });
  await page.getByLabel("Layout zoom", { exact: true }).selectOption("200");
  expect(
    await page.evaluate(
      () => document.documentElement.scrollWidth <= innerWidth,
    ),
  ).toBe(true);
  await page
    .locator(".omr-designer")
    .screenshot({ path: ".local/screenshots/omr-visual-editor-mobile.png" });
  await page.setViewportSize({ width: 1440, height: 1100 });
  await page.getByLabel("Layout zoom", { exact: true }).selectOption("fit");
  await page
    .getByLabel("OMR layout name", { exact: true })
    .fill("Faculty recognition " + paper.id.slice(0, 8));
  const saveResponse = page.waitForResponse(
    (r) =>
      r.url().endsWith("/omr-templates") && r.request().method() === "POST",
  );
  await page
    .getByRole("button", { name: "Save template", exact: true })
    .click();
  const savedResponse = await saveResponse;
  const saved = await savedResponse.json();
  expect(savedResponse.status(), JSON.stringify(saved.message ?? "")).toBe(200);
  await expect(
    page.getByLabel("3. Answer sheets", { exact: true }),
  ).toBeDisabled();
  await page
    .getByLabel("1. Assessment", { exact: true })
    .selectOption(paper.id);
  await page
    .getByLabel("2. OMR template", { exact: true })
    .selectOption(saved.id);
  await expect(page.getByLabel("2. OMR template", { exact: true })).toHaveValue(
    saved.id,
  );
  await expect(
    page.getByLabel("3. Answer sheets", { exact: true }),
  ).toBeEnabled();
  const persisted = await (
    await page.request.get(base + "omr-templates")
  ).json();
  expect(
    persisted.some((t: any) => t.id === saved.id && t.name === saved.name),
  ).toBe(true);
  await page.reload();
  await expect(page.getByLabel("1. Assessment", { exact: true })).toHaveValue(
    paper.id,
  );
  await expect(page.getByLabel("2. OMR template", { exact: true })).toHaveValue(
    saved.id,
  );
  const different = execFileSync(
    "omr/.venv/bin/python",
    [
      "-c",
      `import sys,io
from PIL import Image,ImageDraw,ImageFont
im=Image.open(io.BytesIO(sys.stdin.buffer.read()));d=ImageDraw.Draw(im)
d.ellipse((167,737,193,763),fill='white',outline='#606060',width=2)
d.text((180,750),'A',anchor='mm',font=ImageFont.load_default(size=18),fill='#686868')
d.ellipse((237,737,263,763),fill='#252525',outline='#606060',width=2)
d.rectangle((75,210,350,248),fill='white')
d.text((80,215),'Roll: 000002',font=ImageFont.load_default(size=23),fill='#263846')
im=im.rotate(.8,resample=Image.Resampling.BICUBIC,fillcolor='white')
b=io.BytesIO();im.save(b,format='PNG');sys.stdout.buffer.write(b.getvalue())`,
    ],
    { input: readFileSync(sheet) },
  );
  await page.getByLabel("3. Answer sheets", { exact: true }).setInputFiles([
    {
      name: "first-student.png",
      mimeType: "image/png",
      buffer: readFileSync(sheet),
    },
    { name: "second-student.png", mimeType: "image/png", buffer: different },
  ]);
  await page
    .getByRole("button", { name: "Check answer sheets", exact: true })
    .click();
  await expect(
    page.getByRole("status").filter({ hasText: "2 file(s) accepted" }),
  ).toBeVisible();
  let checks: any[] = [];
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
      { timeout: 60000 },
    )
    .toBe(true);
  const all = checks.flatMap((c) => c.sheets);
  expect(all.map((s) => s.registration_number).sort()).toEqual([
    "000002",
    "123456",
  ]);
  expect(all.find((s) => s.registration_number === "123456")).toMatchObject({
    score: "76",
    maximum: "80",
    state: "needs_review",
    uncertain_questions: [7],
  });
  expect(all.find((s) => s.registration_number === "000002")).toMatchObject({
    score: "71",
    maximum: "80",
    state: "needs_review",
    wrong_questions: [1],
  });
  const table = page.getByRole("region", {
    name: "OMR grades by registration number",
    exact: true,
  });
  for (const registration of ["123456", "000002"]) {
    const row = table.getByRole("row").filter({
      has: page.getByRole("rowheader", {
        name: new RegExp("^" + registration),
      }),
    });
    await row
      .getByRole("button", { name: "Review sheet", exact: true })
      .click();
    const panel = page.getByRole("region", {
      name: "Review scanned sheet",
      exact: true,
    });
    await expect(
      panel.getByLabel("Registration number", { exact: true }),
    ).toHaveValue(registration);
    await panel
      .getByLabel("Sheet review reason", { exact: true })
      .fill(
        "Checked the printed identity and confirmed the deliberately blank question seven.",
      );
    await panel
      .getByLabel("I checked the identity and responses against this sheet.", {
        exact: true,
      })
      .check();
    await panel
      .getByRole("button", { name: "Save reviewed grade", exact: true })
      .click();
    await expect(row).toContainText("Graded");
  }
  await page
    .getByRole("button", { name: "Check answer sheets", exact: true })
    .click();
  await expect(
    page.getByRole("status").filter({ hasText: "2 file(s) accepted" }),
  ).toBeVisible();
  checks = await (
    await page.request.get(base + "omr-checks?paper=" + paper.id)
  ).json();
  expect(checks).toHaveLength(2);
  expect(
    checks.flatMap((c) => c.sheets).every((s) => s.state === "graded"),
  ).toBe(true);
  await page
    .locator(".omrchecker-reader")
    .screenshot({ path: ".local/screenshots/omr-ordered-scanner.png" });
  await page
    .getByRole("region", {
      name: "Registration-based OMR results",
      exact: true,
    })
    .screenshot({ path: ".local/screenshots/omr-registration-grades.png" });
  const shorter = await originalOMRAssessment(page, 3);
  await page.reload();
  await page
    .getByLabel("1. Assessment", { exact: true })
    .selectOption(shorter.paper.id);
  await page
    .getByLabel("2. OMR template", { exact: true })
    .selectOption(saved.id);
  await expect(
    page.getByText(
      "Only the first 3 answer rows are graded for this assessment. The remaining template rows are ignored.",
      { exact: true },
    ),
  ).toBeVisible();
  await page
    .getByLabel("3. Answer sheets", { exact: true })
    .setInputFiles(sheet);
  await page
    .getByRole("button", { name: "Check answer sheets", exact: true })
    .click();
  await expect
    .poll(
      async () => {
        const rows = await (
          await page.request.get(base + "omr-checks?paper=" + shorter.paper.id)
        ).json();
        return rows[0]?.sheets[0]?.score;
      },
      { timeout: 30000 },
    )
    .toBe("12");
  const reused = await (
    await page.request.get(base + "omr-checks?paper=" + shorter.paper.id)
  ).json();
  expect(reused[0].template_id).toBe(saved.id);
  expect(reused[0].sheets[0]).toMatchObject({
    maximum: "12",
    state: "graded",
    correct_questions: [1, 2, 3],
    uncertain_questions: [],
  });
  await navigate(page, "Results");
  await page
    .getByLabel("Assessment to review", { exact: true })
    .selectOption(paper.id);
  await expect(
    page.getByRole("region", {
      name: "OMR grades by registration number",
      exact: true,
    }),
  ).toContainText("000002");
  await page.setViewportSize({ width: 390, height: 844 });
  expect(
    await page.evaluate(
      () => document.documentElement.scrollWidth <= innerWidth,
    ),
  ).toBe(true);
  await page
    .getByRole("region", {
      name: "Registration-based OMR results",
      exact: true,
    })
    .screenshot({
      path: ".local/screenshots/omr-registration-grades-mobile.png",
    });
});
