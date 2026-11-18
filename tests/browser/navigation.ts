import { expect, type Page } from "@playwright/test";
export async function navigate(page: Page, name: string) {
  await page.locator("#workspace-navigation").waitFor({ state: "attached" });
  const menu = page.getByRole("button", { name: "Menu", exact: true });
  await expect(
    page.getByRole("group", { name: "Workspace tools", exact: true }),
  ).toBeEnabled();
  if (await page.evaluate(() => matchMedia("(max-width: 760px)").matches)) {
    if (await menu.count()) await menu.click();
  }
  const nav = page.locator("#workspace-navigation");
  const target = nav.getByRole("button", { name, exact: true });
  if (!(await target.isVisible())) {
    const group = [
      "People",
      "Campuses",
      "Courses",
      "Admissions",
      "Family links",
      "Teaching responsibilities",
      "Enrollment intake",
    ].includes(name)
      ? "People & organization"
      : [
            "Audit ledger",
            "Delivery operations",
            "Institution settings",
          ].includes(name)
        ? "Administration"
        : null;
    if (group) await nav.getByText(group, { exact: true }).click();
  }
  await target.click();
}
