import { test, expect, type Page } from "@playwright/test";
import { readFileSync } from "node:fs";
import { signInStaff } from "./staff-sign-in";
const fixture = JSON.parse(
  readFileSync(".local/academic-browser-access.json", "utf8"),
);
const canvas = (page: Page) =>
  page.getByRole("group", { name: "OMR calibration canvas", exact: true });
async function draw(
  page: Page,
  x1: number,
  y1: number,
  x2: number,
  y2: number,
) {
  const target = canvas(page);
  await target.scrollIntoViewIfNeeded();
  const box = (await target.boundingBox())!;
  await page.mouse.move(box.x + x1 * box.width, box.y + y1 * box.height);
  await page.mouse.down();
  await page.mouse.move(box.x + x2 * box.width, box.y + y2 * box.height, {
    steps: 12,
  });
  await page.mouse.up();
}
async function save(page: Page) {
  const next = page.waitForEvent("download");
  await page
    .getByRole("button", { name: "Save calibration", exact: true })
    .click();
  return readFileSync((await (await next).path())!, "utf8");
}
test("snipping frames drive real OMR pixels, support resize/undo, and round-trip saved calibration", async ({
  page,
  browser,
}) => {
  test.setTimeout(100000);
  const source = await browser.newPage({
    viewport: { width: 800, height: 1000 },
  });
  await source.setContent(
    `<body style="margin:0"><svg width="800" height="1000" style="background:#fff">${[200, 500, 800].map((y, row) => [160, 320, 480, 640].map((x, opt) => `<circle cx="${x}" cy="${y}" r="10" stroke="#888" stroke-width="1.2" fill="${opt === [0, 2, 3][row] ? "#222" : "#fff"}"/>`).join("")).join("")}</svg></body>`,
  );
  const png = await source.screenshot();
  await source.close();
  await page.setViewportSize({ width: 1440, height: 1100 });
  await signInStaff(page);
  await page.goto("/workspace#education/scan");
  await page.getByText("Advanced pixel-reader tools", { exact: true }).click();
  await page.getByText("Visual OMR calibrator", { exact: true }).click();
  await page
    .getByLabel("OMR preview course", { exact: true })
    .selectOption(fixture.course);
  await page.getByLabel("Existing OMR image", { exact: true }).setInputFiles({
    name: "original-calibrator-grid.png",
    mimeType: "image/png",
    buffer: png,
  });
  await expect(canvas(page)).toBeVisible();
  await page.getByLabel("Total questions", { exact: true }).fill("3");
  await page.getByLabel("Calibration zoom", { exact: true }).selectOption("50");
  await draw(page, 0.192, 0.1936, 0.808, 0.8064);
  await expect(
    page.getByRole("status").filter({ hasText: "Created 3 question frames" }),
  ).toBeVisible();
  await page
    .getByRole("button", { name: "Read calibrated answers", exact: true })
    .click();
  await expect(
    page.getByLabel("Detected answer 1", { exact: true }),
  ).toHaveValue("1");
  await expect(
    page.getByLabel("Detected answer 2", { exact: true }),
  ).toHaveValue("3");
  await expect(
    page.getByLabel("Detected answer 3", { exact: true }),
  ).toHaveValue("4");

  await page
    .getByLabel("Selected question frame", { exact: true })
    .selectOption("2");
  await draw(page, 0.192, 0.4, 0.808, 0.425);
  await expect(
    page.getByLabel("Detected answer 2", { exact: true }),
  ).not.toBeVisible();
  await page
    .getByRole("button", { name: "Read calibrated answers", exact: true })
    .click();
  await expect(
    page.getByLabel("Detected answer 2", { exact: true }),
  ).toHaveValue("blank");

  await page
    .getByRole("button", { name: "Move / resize", exact: true })
    .click();
  await draw(page, 0.5, 0.4125, 0.5, 0.5);
  await page
    .getByRole("button", { name: "Read calibrated answers", exact: true })
    .click();
  await expect(
    page.getByLabel("Detected answer 2", { exact: true }),
  ).toHaveValue("3");
  const saved = await save(page),
    c = JSON.parse(saved);
  expect(c.frames).toHaveLength(1);
  expect(c.frames[0].number).toBe(2);
  expect(c.frames[0].rect.y + c.frames[0].rect.height / 2).toBeCloseTo(0.5, 2);
  const width = Number(
    await page
      .getByLabel("Selected frame width percent", { exact: true })
      .inputValue(),
  );
  const handle = page.locator(".omr-handle-se");
  await handle.scrollIntoViewIfNeeded();
  const box = (await handle.boundingBox())!;
  await page.mouse.move(box.x + box.width / 2, box.y + box.height / 2);
  await page.mouse.down();
  await page.mouse.move(box.x + box.width / 2 + 2, box.y + box.height / 2 + 2, {
    steps: 4,
  });
  await page.mouse.up();
  await expect(
    page
      .getByRole("status")
      .filter({ hasText: "Keep this frame inside the answer area" }),
  ).toBeVisible();
  await handle.scrollIntoViewIfNeeded();
  const current = (await handle.boundingBox())!;
  await page.mouse.move(
    current.x + current.width / 2,
    current.y + current.height / 2,
  );
  await page.mouse.down();
  await page.mouse.move(
    current.x + current.width / 2 - 2,
    current.y + current.height / 2 + 2,
    {
      steps: 4,
    },
  );
  await page.mouse.up();
  await expect
    .poll(async () =>
      Number(
        await page
          .getByLabel("Selected frame width percent", { exact: true })
          .inputValue(),
      ),
    )
    .toBeLessThan(width);
  await page
    .getByRole("button", { name: "Undo frame change", exact: true })
    .click();
  await expect(
    page.getByLabel("Selected frame width percent", { exact: true }),
  ).not.toBeVisible();
  await page
    .getByRole("button", { name: "Draw answer area", exact: true })
    .click();
  const before = await save(page);
  const bounds = (await canvas(page).boundingBox())!;
  await page.mouse.move(
    bounds.x + bounds.width * 0.25,
    bounds.y + bounds.height * 0.3,
  );
  await page.mouse.down();
  await page.mouse.move(
    bounds.x + bounds.width * 0.7,
    bounds.y + bounds.height * 0.7,
  );
  await canvas(page).press("Escape");
  await page.mouse.up();
  expect(await save(page)).toBe(before);
  await page.getByLabel("Load OMR calibration", { exact: true }).setInputFiles({
    name: "saved-calibration.json",
    mimeType: "application/json",
    buffer: Buffer.from(saved),
  });
  await expect(
    page.getByRole("status").filter({ hasText: "Calibration loaded" }),
  ).toBeVisible();
  await page
    .getByRole("button", { name: "Read calibrated answers", exact: true })
    .click();
  await expect(
    page.getByLabel("Detected answer 2", { exact: true }),
  ).toHaveValue("3");
  await page
    .locator(".omr-calibrator")
    .screenshot({ path: ".local/screenshots/omr-calibrator-desktop.png" });
  await page.setViewportSize({ width: 390, height: 844 });
  await expect(
    page.getByRole("button", { name: "Draw question frame", exact: true }),
  ).toBeVisible();
  expect(
    await page.evaluate(
      () => document.documentElement.scrollWidth > innerWidth,
    ),
  ).toBe(false);
  await page
    .locator(".omr-calibrator")
    .screenshot({ path: ".local/screenshots/omr-calibrator-mobile.png" });
});
