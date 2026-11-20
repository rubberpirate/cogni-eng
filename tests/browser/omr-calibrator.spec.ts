import { test, expect, type Page } from "@playwright/test";
import { existsSync, readFileSync } from "node:fs";
import { signInStaff } from "./staff-sign-in";
import { navigate } from "./navigation";
const fixture = JSON.parse(
  readFileSync(".local/academic-browser-access.json", "utf8"),
);
const sample = "test-papers/OMR-CTET-SHEET-Sample.jpg";
const evidence = ".local/live-assessment/omr-OMR-CTET-SHEET-Sample.jpg.json";
const surface = (page: Page) =>
  page.getByRole("group", { name: "Template calibration canvas", exact: true });
async function draw(
  page: Page,
  x: number,
  y: number,
  right: number,
  bottom: number,
) {
  await page.locator(".omr-sheet-stage").scrollIntoViewIfNeeded();
  const b = (await surface(page).boundingBox())!;
  await page.mouse.move(b.x + x * b.width, b.y + y * b.height);
  await page.mouse.down();
  await page.mouse.move(b.x + right * b.width, b.y + bottom * b.height, {
    steps: 50,
  });
  await page.mouse.up();
}
test("draw then label frames on the whole CTET sheet without premature circles", async ({
  page,
}) => {
  test.skip(
    !existsSync(sample) || !existsSync(evidence),
    "Owner's private sheet and independently checked local reference required.",
  );
  test.setTimeout(140000);
  const known = JSON.parse(readFileSync(evidence, "utf8"));
  await page.setViewportSize({ width: 1440, height: 1000 });
  await signInStaff(page);
  await expect(
    page.getByRole("heading", { name: "Draft assessment", exact: true }),
  ).toBeVisible();
  await navigate(page, "Upload OMR answer sheets");
  await expect(
    page.getByRole("group", { name: "OMR calibration canvas", exact: true }),
  ).toHaveCount(0);
  await page
    .getByLabel("OMRChecker student batch", { exact: true })
    .selectOption(fixture.course);
  await page
    .getByLabel("OMRChecker sheet image", { exact: true })
    .setInputFiles(sample);
  await page
    .getByRole("button", { name: "Expand editor", exact: true })
    .click();
  const dialog = page.getByRole("dialog");
  await expect(dialog).toBeVisible();
  await expect
    .poll(() =>
      surface(page).evaluate((svg) => {
        const image = svg.previousElementSibling as HTMLImageElement;
        const stage = svg.closest(".omr-sheet-stage")!.getBoundingClientRect(),
          r = image.getBoundingClientRect();
        return (
          image.complete &&
          image.naturalWidth > 0 &&
          r.top >= stage.top &&
          r.bottom <= stage.bottom &&
          r.left >= stage.left &&
          r.right <= stage.right
        );
      }),
    )
    .toBe(true);
  await expect(surface(page).locator("ellipse")).toHaveCount(0);
  await draw(page, 0.03, 0.122, 0.216, 0.15);
  await expect(surface(page).locator("ellipse")).toHaveCount(0);
  await page
    .getByLabel("Field label", { exact: true })
    .selectOption("roll_number");
  await page.getByRole("button", { name: "Save frame", exact: true }).click();
  await expect(
    page
      .getByText("Registration number", { exact: true })
      .filter({ visible: true }),
  ).toHaveCount(2);
  for (const [index, block] of known.calibration.blocks.entries()) {
    await draw(
      page,
      block.top_left.x - 0.014,
      block.top_left.y - 0.012,
      block.bottom_right.x + 0.014,
      block.bottom_right.y + 0.012,
    );
    await expect(surface(page).locator("ellipse")).toHaveCount(0);
    await page
      .getByLabel("Field label", { exact: true })
      .selectOption("answers");
    await page
      .getByLabel("Option labels", { exact: true })
      .selectOption("numbers");
    await expect(surface(page).locator("ellipse")).toHaveCount(0);
    await page
      .getByRole("button", { name: "Mark first two rows", exact: true })
      .click();
    await expect(page.getByLabel("Layout zoom", { exact: true })).toHaveValue(
      "block",
    );
    for (let row = 0; row < 2; row++)
      for (let col = 0; col < 4; col++) {
        const t = row / 29,
          u = col / 3;
        const left = {
          x: block.top_left.x + (block.bottom_left.x - block.top_left.x) * t,
          y: block.top_left.y + (block.bottom_left.y - block.top_left.y) * t,
        };
        const right = {
          x: block.top_right.x + (block.bottom_right.x - block.top_right.x) * t,
          y: block.top_right.y + (block.bottom_right.y - block.top_right.y) * t,
        };
        await page.locator(".omr-sheet-stage").scrollIntoViewIfNeeded();
        const b = (await surface(page).boundingBox())!;
        await page.mouse.click(
          b.x + (left.x + (right.x - left.x) * u) * b.width,
          b.y + (left.y + (right.y - left.y) * u) * b.height,
        );
      }
    await expect(
      page.getByLabel("Questions in this box", { exact: true }),
    ).toHaveValue("30");
    await expect(
      page.getByLabel("Block first question", { exact: true }),
    ).toHaveValue(String(index * 30 + 1));
    await page.getByRole("button", { name: "Save frame", exact: true }).click();
    await expect(surface(page).locator("ellipse")).toHaveCount(0);
    await expect(page.getByLabel("Layout zoom", { exact: true })).toHaveValue(
      "fit",
    );
  }
  await page
    .getByLabel("OMR layout name", { exact: true })
    .fill("CTET 150-question layout");
  await dialog.screenshot({
    path: ".local/screenshots/omr-ctet-frame-editor.png",
  });
  await page
    .getByRole("button", { name: "Close expanded editor", exact: true })
    .click();
  await expect(dialog).not.toBeVisible();
  await page.setViewportSize({ width: 390, height: 844 });
  await expect
    .poll(() =>
      page.evaluate(() => document.documentElement.scrollWidth <= innerWidth),
    )
    .toBe(true);
  await page
    .locator(".omr-designer")
    .screenshot({ path: ".local/screenshots/omr-ctet-frame-mobile.png" });
  await page.setViewportSize({ width: 1440, height: 1000 });
  await page
    .getByRole("button", { name: "Use this layout", exact: true })
    .click();
  await expect(
    page.getByText(
      "CTET 150-question layout · 150 questions · 5 answer blocks",
      { exact: true },
    ),
  ).toBeVisible({ timeout: 30000 });
  const download = page.waitForEvent("download");
  await page
    .getByRole("button", { name: "Save reusable layout", exact: true })
    .click();
  const saved = JSON.parse(
    readFileSync((await (await download).path())!, "utf8"),
  );
  expect(Object.keys(saved.template.fieldBlocks)).toHaveLength(150);
  expect(saved.template.fieldBlocks.answer_1.bubbleValues).toEqual([
    "1",
    "2",
    "3",
    "4",
  ]);
  if (process.env.RUN_OMRCHECKER_TESTS === "1") {
    const response = page.waitForResponse(
      (r) =>
        r.url().endsWith("/omrchecker-preview") &&
        r.request().method() === "POST",
    );
    await page
      .getByRole("button", { name: "Read OMRChecker sheet", exact: true })
      .click();
    const r = await response,
      result = await r.json();
    expect(r.status(), JSON.stringify(result.message ?? "")).toBe(200);
    const expected = Object.fromEntries(
      known.expected.map((v: number, i: number) => [
        `q${i + 1}`,
        v ? String(v) : "",
      ]),
    );
    expect(result.responses).toEqual(expected);
    await page
      .locator(".omrchecker-reader")
      .screenshot({ path: ".local/screenshots/omr-ctet-native-result.png" });
  }
});
