import { test, expect } from "@playwright/test";
import { readFileSync } from "node:fs";
const creds = readFileSync(".local/demo-access.txt", "utf8");
const email = creds.match(/Email: (.+)/)![1];
const password = creds.match(/Password: (.+)/)![1];
test("owner organizes campus, course, admission and family with persistent audit", async ({
  page,
}) => {
  await page.goto("/sign-in");
  await page.getByLabel("Email address", { exact: true }).fill(email);
  await page.getByLabel("Password", { exact: true }).fill(password);
  await page.getByRole("button", { name: "Sign in", exact: true }).click();
  await expect(
    page.getByRole("heading", { name: "People", exact: true }),
  ).toBeVisible();
  const suffix = Date.now().toString().slice(-7);
  const campus = "Test Campus " + suffix;
  await page.getByRole("button", { name: "Campuses", exact: true }).click();
  const campusForm = page
    .getByRole("heading", { name: "Create · Campuses" })
    .locator("..")
    .locator("form");
  await campusForm.getByLabel("Name", { exact: true }).fill(campus);
  await campusForm.getByLabel("Code", { exact: true }).fill("T" + suffix);
  await campusForm
    .getByLabel("Address", { exact: true })
    .fill("Synthetic address");
  await campusForm.getByRole("button", { name: "Save record" }).click();
  await expect(
    page.getByRole("cell", { name: campus, exact: true }),
  ).toBeVisible();
  await page.getByRole("button", { name: "Courses", exact: true }).click();
  const courseForm = page
    .getByRole("heading", { name: "Create · Courses" })
    .locator("..")
    .locator("form");
  await courseForm
    .getByLabel("Name", { exact: true })
    .fill("Science " + suffix);
  await courseForm
    .getByLabel("Campus", { exact: true })
    .selectOption({ label: campus });
  await courseForm.getByLabel("Academic year").fill("2026–27");
  await courseForm.getByRole("button", { name: "Save record" }).click();
  await expect(
    page.getByRole("cell", { name: "Science " + suffix, exact: true }),
  ).toBeVisible();
  await page.getByRole("button", { name: "People", exact: true }).click();
  const form = page
    .getByRole("heading", { name: "Invite a person" })
    .locator("..")
    .locator("form");
  await form.getByLabel("Name", { exact: true }).fill("Learner " + suffix);
  await form
    .getByLabel("Email address")
    .fill("learner-" + suffix + "@demonstration.invalid");
  await form
    .getByLabel("Campus", { exact: true })
    .selectOption({ label: campus });
  await form.getByRole("button", { name: "Save record" }).click();
  await expect(
    page.getByRole("cell", { name: "Learner " + suffix, exact: true }),
  ).toBeVisible();
  await page.getByRole("button", { name: "Admissions", exact: true }).click();
  const admission = page
    .getByRole("heading", { name: "Create · Admissions" })
    .locator("..")
    .locator("form");
  await admission
    .getByLabel("Campus", { exact: true })
    .selectOption({ label: campus });
  await admission
    .getByLabel("Learner", { exact: true })
    .selectOption({
      label:
        "Learner " + suffix + " · learner-" + suffix + "@demonstration.invalid",
    });
  await admission
    .getByLabel("Course", { exact: true })
    .selectOption({ label: "Science " + suffix });
  await admission.getByLabel("Start date").fill("2026-11-17");
  await admission.getByRole("button", { name: "Save record" }).click();
  await expect(
    page.getByRole("cell", { name: "Learner " + suffix, exact: true }),
  ).toBeVisible();
  await page.reload();
  await expect(
    page.getByRole("heading", { name: "People", exact: true }),
  ).toBeVisible();
  await page.getByRole("button", { name: "Audit ledger", exact: true }).click();
  await expect(
    page.getByRole("cell", { name: "learner.enrolled", exact: true }).first(),
  ).toBeVisible();
  await page.screenshot({
    path: ".local/screenshots/audit-desktop.png",
    fullPage: true,
  });
  await page
    .getByRole("button", { name: "Access and security", exact: true })
    .click();
  await expect(
    page.getByRole("heading", { name: "Your active sessions" }),
  ).toBeVisible();
  await page.getByRole("button", { name: "Sign out", exact: true }).click();
  await expect(page).toHaveURL(/sign-in/);
  await page.goto("/workspace");
  await expect(page).toHaveURL(/sign-in/);
});
test("narrow layout and keyboard sign-in retain meaningful states", async ({
  page,
}) => {
  await page.setViewportSize({ width: 390, height: 844 });
  await page.goto("/sign-in");
  await page.getByLabel("Email address", { exact: true }).fill(email);
  await page.getByLabel("Password", { exact: true }).fill(password);
  await page.getByLabel("Password", { exact: true }).press("Enter");
  await expect(
    page.getByRole("heading", { name: "People", exact: true }),
  ).toBeVisible();
  expect(
    await page.evaluate(
      () => document.documentElement.scrollWidth <= innerWidth,
    ),
  ).toBe(true);
  await page.getByLabel("Language", { exact: true }).selectOption("hi");
  await expect(
    page.getByRole("heading", { name: "लोग", exact: true }),
  ).toBeVisible();
  await page.screenshot({
    path: ".local/screenshots/people-mobile-hi.png",
    fullPage: true,
  });
});
