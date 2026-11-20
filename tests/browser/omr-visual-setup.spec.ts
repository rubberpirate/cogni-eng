import { test, expect, type Page } from "@playwright/test";
import { readFileSync } from "node:fs";
import { resolve } from "node:path";
import { execFileSync } from "node:child_process";
import { signInStaff } from "./staff-sign-in";
import { navigate } from "./navigation";
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
async function read(page: Page) {
  const response = page.waitForResponse(
    (r) =>
      r.url().endsWith("/omrchecker-preview") &&
      r.request().method() === "POST",
  );
  await page
    .getByRole("button", { name: "Read OMRChecker sheet", exact: true })
    .click();
  const r = await response;
  const result = await r.json();
  expect(r.status(), JSON.stringify(result.error ?? result.message ?? "")).toBe(
    200,
  );
  return result;
}
test("name/roll snips and two clicked rows build a reusable native OMRChecker layout", async ({
  page,
}) => {
  test.setTimeout(140000);
  await page.setViewportSize({ width: 1440, height: 1100 });
  await signInStaff(page);
  await expect(
    page.getByRole("heading", { name: "Draft assessment", exact: true }),
  ).toBeVisible();
  await navigate(page, "Upload OMR answer sheets");
  await page
    .getByLabel("OMRChecker student batch", { exact: true })
    .selectOption(fixture.course);
  await page
    .getByLabel("OMRChecker sheet image", { exact: true })
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
    .fill("Visual test layout");
  await page
    .getByRole("button", { name: "Use this layout", exact: true })
    .click();
  await expect(
    page.getByText("Visual test layout · 20 questions · 2 answer blocks", {
      exact: true,
    }),
  ).toBeVisible();
  await expect(
    page.getByLabel("OMRChecker template JSON", { exact: true }),
  ).toHaveCount(0);
  const download = page.waitForEvent("download");
  await page
    .getByRole("button", { name: "Save reusable layout", exact: true })
    .click();
  const saved = readFileSync((await (await download).path())!);
  const profile = JSON.parse(saved.toString());
  expect(profile.layout.blocks.map((b: any) => b.points.length)).toEqual([
    8, 8,
  ]);
  expect(profile.reference.data.length).toBeGreaterThan(1000);
  expect(profile.template.preProcessors[0].name).toBe("FeatureBasedAlignment");
  const result = await read(page);
  const expected = JSON.parse(
    readFileSync(
      resolve("apps/web/public/samples/omrchecker/expected-answers.json"),
      "utf8",
    ),
  ).responses;
  delete expected.Roll;
  expect(result.responses).toEqual(expected);
  expect(result.identity.find((r: any) => r.kind === "roll_number").text).toBe(
    "123456",
  );
  expect(
    result.identity.find((r: any) => r.kind === "student_name").text,
  ).toContain("Synthetic Test Learner");
  await expect(
    page.getByLabel("Detected roll number", { exact: true }),
  ).toHaveValue("123456");
  await page.getByLabel("Detected roll number", { exact: true }).fill("654321");
  await expect(
    page.getByText("Manually reviewed", { exact: true }),
  ).toBeVisible();
  const answerDownload = page.waitForEvent("download");
  await page
    .getByRole("button", { name: "Download detected answers", exact: true })
    .click();
  const reviewed = JSON.parse(
    readFileSync((await (await answerDownload).path())!, "utf8"),
  );
  expect(
    reviewed.identity.find((r: any) => r.kind === "roll_number").text,
  ).toBe("654321");
  await page.getByLabel("Detected roll number", { exact: true }).fill("123456");
  await page
    .locator(".omrchecker-reader")
    .screenshot({ path: ".local/screenshots/omr-visual-native-desktop.png" });
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
im=im.rotate(.8,resample=Image.Resampling.BICUBIC,fillcolor='white')
b=io.BytesIO();im.save(b,format='PNG');sys.stdout.buffer.write(b.getvalue())`,
    ],
    { input: readFileSync(sheet) },
  );
  await page
    .getByLabel("OMRChecker sheet image", { exact: true })
    .setInputFiles({
      name: "another-filled-sheet.png",
      mimeType: "image/png",
      buffer: different,
    });
  await expect(
    page.getByText("Visual test layout · 20 questions · 2 answer blocks", {
      exact: true,
    }),
  ).toBeVisible();
  expect((await read(page)).responses).toEqual({ ...expected, q1: "B" });
  await page.reload();
  await navigate(page, "Upload OMR answer sheets");
  await page
    .getByLabel("Load saved OMR layout", { exact: true })
    .setInputFiles({
      name: "visual.omr.json",
      mimeType: "application/json",
      buffer: saved,
    });
  await page
    .getByLabel("OMRChecker student batch", { exact: true })
    .selectOption(fixture.course);
  await page
    .getByLabel("OMRChecker sheet image", { exact: true })
    .setInputFiles({
      name: "another-filled-sheet.png",
      mimeType: "image/png",
      buffer: different,
    });
  expect((await read(page)).responses).toEqual({ ...expected, q1: "B" });
  await page.setViewportSize({ width: 390, height: 844 });
  expect(
    await page.evaluate(
      () => document.documentElement.scrollWidth > innerWidth,
    ),
  ).toBe(false);
  await page
    .locator(".omrchecker-reader")
    .screenshot({ path: ".local/screenshots/omr-visual-native-mobile.png" });
});
