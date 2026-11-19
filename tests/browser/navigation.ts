import { expect, type Page } from "@playwright/test";
export async function navigate(page: Page, name: string) {
  await page.locator("#workspace-navigation").waitFor({ state: "attached" });
  const legacy: Record<string, string> = {
    Home: "home",
    "Question bank": "curriculum",
    "Upload paper": "import",
    "Enter questions manually": "author",
    "Create questions with AI": "generate",
    Assessments: "papers",
  };
  if (legacy[name]) {
    await expect(
      page.getByRole("group", { name: "Workspace tools", exact: true }),
    ).toBeEnabled();
    await page.evaluate((route) => {
      const query = location.hash.split("?")[1];
      const params = new URLSearchParams(query);
      params.delete("draft");
      location.hash = `education/${route}${params.size ? "?" + params : ""}`;
    }, legacy[name]);
    return;
  }
  name =
    (
      {
        "Scan & check": "Upload OMR answer sheets",
        "Results & review": "Results",
      } as Record<string, string>
    )[name] ?? name;
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
      : ["Background tasks", "Try sample files"].includes(name)
        ? "Tools & resources"
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
