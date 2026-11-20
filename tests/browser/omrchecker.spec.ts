import { test, expect } from "@playwright/test";
import { readFileSync } from "node:fs";
import { resolve } from "node:path";
import { signInStaff } from "./staff-sign-in";
import { navigate } from "./navigation";
test.skip(
  process.env.RUN_OMRCHECKER_TESTS !== "1",
  "Opt-in: actual installed OMRChecker checkout/runtime, no HTTP recognition fixture.",
);
const fixture = JSON.parse(
  readFileSync(".local/academic-browser-access.json", "utf8"),
);
const samples = resolve("apps/web/public/samples/omrchecker");
test("native OMRChecker template, real upload, overlay and answer download agree with the filled sample", async ({
  page,
}) => {
  test.setTimeout(100000);
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
    .setInputFiles(resolve(samples, "filled-sheet.png"));
  await page
    .getByLabel("OMRChecker template JSON", { exact: true })
    .setInputFiles(resolve(samples, "template.json"));
  const response = page.waitForResponse(
    (r) =>
      r.url().endsWith("/omrchecker-preview") &&
      r.request().method() === "POST",
  );
  await page
    .getByRole("button", { name: "Read OMRChecker sheet", exact: true })
    .click();
  const result = await response;
  expect(result.status()).toBe(200);
  const native = await result.json();
  expect(native.engine).toBe("OMRChecker");
  expect(native.revision).toMatch(/^[a-f0-9]{40}$/);
  expect(native.responses).toEqual(
    JSON.parse(readFileSync(resolve(samples, "expected-answers.json"), "utf8"))
      .responses,
  );
  await expect(
    page.getByRole("heading", { name: /26 fields read · 2 need review/ }),
  ).toBeVisible();
  const table = page.getByRole("table", { name: "Detected template fields" });
  await expect(
    table.getByRole("row").filter({
      has: page.getByRole("rowheader", { name: "q7", exact: true }),
    }),
  ).toContainText("Blank");
  await expect(
    table.getByRole("row").filter({
      has: page.getByRole("rowheader", { name: "q12", exact: true }),
    }),
  ).toContainText("A + C");
  const overlay = page.getByRole("img", {
    name: "OMRChecker aligned sheet with detected bubbles marked",
  });
  await expect
    .poll(() =>
      overlay.evaluate(
        (img: HTMLImageElement) => img.complete && img.naturalWidth > 0,
      ),
    )
    .toBe(true);
  const download = page.waitForEvent("download");
  await page
    .getByRole("button", { name: "Download detected answers", exact: true })
    .click();
  const file = await download;
  const report = JSON.parse(readFileSync((await file.path())!, "utf8"));
  expect(report.responses).toEqual(native.responses);
  expect(report).not.toHaveProperty("overlay");
  const me = await (await page.request.get("/api/v1/me")).json();
  const missingMarker = await page.request.post(
    `/api/v1/institutions/${fixture.org}/education/omrchecker-preview`,
    {
      headers: { "X-CSRF-Token": me.csrf, Origin: "http://localhost:3000" },
      multipart: {
        course_id: fixture.course,
        file: {
          name: "sheet.png",
          mimeType: "image/png",
          buffer: readFileSync(resolve(samples, "filled-sheet.png")),
        },
        template: JSON.stringify({
          ...JSON.parse(
            readFileSync(resolve(samples, "template.json"), "utf8"),
          ),
          preProcessors: [
            {
              name: "CropOnMarkers",
              options: { relativePath: "../../not-an-upload" },
            },
          ],
        }),
      },
    },
  );
  expect(missingMarker.status()).toBe(400);
  expect((await missingMarker.json()).message).toContain("marker image");
  await page
    .locator(".omrchecker-reader")
    .screenshot({ path: ".local/screenshots/omrchecker-reader-desktop.png" });
  await page.setViewportSize({ width: 390, height: 844 });
  expect(
    await page.evaluate(
      () => document.documentElement.scrollWidth > innerWidth,
    ),
  ).toBe(false);
  await page
    .locator(".omrchecker-reader")
    .screenshot({ path: ".local/screenshots/omrchecker-reader-mobile.png" });
});
