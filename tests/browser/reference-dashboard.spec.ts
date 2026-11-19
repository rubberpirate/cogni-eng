import { test, expect } from "@playwright/test";
import { readFileSync } from "node:fs";
import { navigate } from "./navigation";
const creds = readFileSync(
  process.env.E2E_ACCESS_FILE ?? ".local/browser-access.txt",
  "utf8",
);
const fixture = JSON.parse(
  readFileSync(".local/academic-browser-access.json", "utf8"),
);
test("reference dashboard renders live counts, searches saved papers and opens real tools at all screen sizes", async ({
  page,
}) => {
  await page.goto("/sign-in");
  await page
    .getByLabel("Email address", { exact: true })
    .fill(creds.match(/Email: (.+)/)![1]);
  await page
    .getByLabel("Password", { exact: true })
    .fill(creds.match(/Password: (.+)/)![1]);
  await page.getByRole("button", { name: "Sign in", exact: true }).click();
  await expect(page.locator(".dashboard-counts")).toBeVisible();
  const papers = await (
    await page.request.get(
      `/api/v1/institutions/${fixture.org}/education/papers?campus=`,
    )
  ).json();
  const topics = await (
    await page.request.get(
      `/api/v1/institutions/${fixture.org}/education/topics?campus=`,
    )
  ).json();
  const stats = page.locator(".dashboard-stat .stat-count");
  await expect(stats).toHaveText([
    String(papers.length),
    String(
      papers.filter((p: { state: string }) => p.state === "published").length,
    ),
    String(
      papers.filter((p: { state: string }) => p.state !== "published").length,
    ),
    String(topics.length),
  ]);
  for (const width of [1440, 1024, 768, 390]) {
    await page.setViewportSize({ width, height: 900 });
    expect(
      await page.evaluate(
        () => document.documentElement.scrollWidth <= innerWidth,
      ),
    ).toBe(true);
    await expect(
      page.getByRole("button", { name: /01 Upload a paper/ }),
    ).toBeVisible();
    await page.screenshot({
      path: `.local/screenshots/dashboard-reference-${width}.png`,
      fullPage: true,
    });
  }
  const search = page.getByLabel("Search assessments", { exact: true });
  await search.fill("no-matching-paper-unique");
  await expect(
    page.getByText("No assessments match your search.", { exact: true }),
  ).toBeVisible();
  if (papers.length) {
    await search.fill(papers[0].reference);
    await expect(page.locator(".dashboard-paper")).toHaveCount(1);
    await page.locator(".dashboard-paper").click();
    await expect(page).toHaveURL(
      new RegExp(`education/papers.*paper=${papers[0].id}`),
    );
    await navigate(page, "Home");
  }
  await page.setViewportSize({ width: 1440, height: 900 });
  await page.getByText("Display & language", { exact: true }).click();
  await page.getByLabel("Appearance", { exact: true }).selectOption("dark");
  await page.getByText("Display & language", { exact: true }).click();
  await page.screenshot({
    path: ".local/screenshots/dashboard-reference-dark.png",
    fullPage: true,
  });
  await page
    .getByRole("button", { name: "Upload OMR sheets", exact: true })
    .click();
  await expect(
    page.getByRole("heading", { name: "Scan & check", exact: true }),
  ).toBeVisible();
});
