import { expect, type Page, type BrowserContext } from "@playwright/test";
import { readFileSync } from "node:fs";
const credentials = readFileSync(
  process.env.E2E_ACCESS_FILE ?? ".local/browser-access.txt",
  "utf8",
);
let cookies: Parameters<BrowserContext["addCookies"]>[0] = [];
// Reuse one synthetic staff session so a full workflow suite does not exercise the
// account's intentional 12-attempt sign-in throttle with successful logins.
export async function signInStaff(page: Page) {
  if (cookies.length) {
    await page.context().addCookies(cookies);
    const me = await page.request.get("/api/v1/me");
    if (me.ok()) {
      await page.goto("/workspace");
      await expect(
        page.getByRole("group", { name: "Workspace tools", exact: true }),
      ).toBeEnabled();
      return;
    }
    await page.context().clearCookies();
    cookies = [];
  }
  await page.goto("/sign-in");
  await page
    .getByLabel("Email address", { exact: true })
    .fill(credentials.match(/Email: (.+)/)![1]);
  await page
    .getByLabel("Password", { exact: true })
    .fill(credentials.match(/Password: (.+)/)![1]);
  await page.getByRole("button", { name: "Sign in", exact: true }).click();
  await expect(
    page.getByRole("group", { name: "Workspace tools", exact: true }),
  ).toBeEnabled();
  cookies = await page.context().cookies();
}
