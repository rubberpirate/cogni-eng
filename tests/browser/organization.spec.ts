import { test, expect } from "@playwright/test";
import { readFileSync } from "node:fs";
const creds = readFileSync(
  process.env.E2E_ACCESS_FILE ?? ".local/demo-access.txt",
  "utf8",
);
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
  await admission.getByLabel("Learner", { exact: true }).selectOption({
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
  await page.getByRole("button", { name: "People", exact: true }).click();
  const guardianForm = page
    .getByRole("heading", { name: "Invite a person" })
    .locator("..")
    .locator("form");
  await guardianForm
    .getByLabel("Name", { exact: true })
    .fill("Guardian " + suffix);
  await guardianForm
    .getByLabel("Email address")
    .fill("guardian-" + suffix + "@demonstration.invalid");
  await guardianForm
    .getByLabel("Responsibility", { exact: true })
    .selectOption("guardian");
  await guardianForm
    .getByLabel("Campus", { exact: true })
    .selectOption({ label: campus });
  await guardianForm.getByRole("button", { name: "Save record" }).click();
  await expect(
    page.getByRole("cell", { name: "Guardian " + suffix, exact: true }),
  ).toBeVisible();
  await page.getByRole("button", { name: "Family links", exact: true }).click();
  const familyForm = page
    .getByRole("heading", { name: "Create · Family links" })
    .locator("..")
    .locator("form");
  await familyForm
    .getByLabel("Campus", { exact: true })
    .selectOption({ label: campus });
  await familyForm.getByLabel("Guardian", { exact: true }).selectOption({
    label:
      "Guardian " + suffix + " · guardian-" + suffix + "@demonstration.invalid",
  });
  await familyForm.getByLabel("Learner", { exact: true }).selectOption({
    label:
      "Learner " + suffix + " · learner-" + suffix + "@demonstration.invalid",
  });
  await familyForm
    .getByLabel("Relationship", { exact: true })
    .fill("Synthetic guardian");
  await familyForm.getByRole("button", { name: "Save record" }).click();
  const familyRow = page.getByRole("row").filter({
    has: page.getByRole("cell", { name: "Guardian " + suffix, exact: true }),
  });
  await expect(
    familyRow.getByRole("cell", { name: "approved", exact: true }),
  ).toBeVisible();
  page.once("dialog", (d) => d.accept());
  await familyRow
    .getByRole("button", { name: "Remove link", exact: true })
    .click();
  await expect(
    familyRow.getByRole("cell", { name: "removed", exact: true }),
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

test("profile and institution edits persist; campus edit uses optimistic version", async ({
  page,
}) => {
  await page.goto("/sign-in");
  await page.getByLabel("Email address", { exact: true }).fill(email);
  await page.getByLabel("Password", { exact: true }).fill(password);
  await page.getByRole("button", { name: "Sign in", exact: true }).click();
  await expect(
    page.getByRole("heading", { name: "People", exact: true }),
  ).toBeVisible();
  await page.getByRole("button", { name: "My account", exact: true }).click();
  const profile = page
    .getByRole("heading", { name: "Your profile" })
    .locator("..")
    .locator("form");
  await profile
    .getByLabel("Emergency contact", { exact: true })
    .fill("Synthetic emergency contact");
  await profile
    .getByRole("button", { name: "Save profile", exact: true })
    .click();
  await expect(
    page.getByRole("status").filter({ hasText: "Profile saved." }),
  ).toBeVisible();
  await page.reload();
  await page.getByRole("button", { name: "My account", exact: true }).click();
  await expect(
    page.getByLabel("Emergency contact", { exact: true }),
  ).toHaveValue("Synthetic emergency contact");
  await page
    .getByRole("button", { name: "Institution settings", exact: true })
    .click();
  const settings = page
    .getByRole("heading", { name: "Institution configuration" })
    .locator("..")
    .locator("form");
  await settings
    .getByLabel("Support email", { exact: true })
    .fill("support@demonstration.invalid");
  await settings
    .getByRole("button", { name: "Save configuration", exact: true })
    .click();
  await expect(
    page.getByRole("status").filter({ hasText: "Configuration saved." }),
  ).toBeVisible();
  await page.getByRole("button", { name: "Campuses", exact: true }).click();
  const row = page.getByRole("row").filter({
    has: page.getByRole("cell", { name: "River Campus", exact: true }),
  });
  await row.getByText("Edit campus", { exact: true }).click();
  await row
    .getByLabel("Address", { exact: true })
    .fill("Updated synthetic address");
  await row.getByRole("button", { name: "Save campus", exact: true }).click();
  await expect(
    row.getByRole("cell", { name: "Updated synthetic address", exact: true }),
  ).toBeVisible();
});

test("learner registers, verifies captured mail and receives approved placement once", async ({
  page,
  browser,
  request,
}) => {
  await page.goto("/sign-in");
  await page.getByLabel("Email address", { exact: true }).fill(email);
  await page.getByLabel("Password", { exact: true }).fill(password);
  await page.getByRole("button", { name: "Sign in", exact: true }).click();
  await expect(
    page.getByRole("heading", { name: "People", exact: true }),
  ).toBeVisible();
  await page
    .getByRole("button", { name: "Enrollment intake", exact: true })
    .click();
  const codeForm = page
    .getByRole("heading", { name: "Enrollment codes", exact: true })
    .locator("..")
    .locator("form")
    .first();
  await codeForm
    .getByLabel("Campus", { exact: true })
    .selectOption({ label: "River Campus" });
  await codeForm
    .getByLabel("Label", { exact: true })
    .fill("Browser intake " + Date.now());
  await codeForm
    .getByRole("button", { name: "Create enrollment code", exact: true })
    .click();
  const code = (await page.locator("code.secret").innerText()).trim();
  const context = await browser.newContext();
  const learner = await context.newPage();
  const address = "registration-" + Date.now() + "@demonstration.invalid";
  try {
    await learner.goto("http://localhost:3000/register");
    await learner
      .getByLabel("Name", { exact: true })
      .fill("Registration Example");
    await learner.getByLabel("Email address", { exact: true }).fill(address);
    await learner
      .getByLabel("New password", { exact: true })
      .fill("a browser registration phrase 3!");
    await learner.getByLabel("Enrollment code", { exact: true }).fill(code);
    await learner.getByRole("checkbox").check();
    await learner
      .getByRole("button", { name: "Request enrollment", exact: true })
      .click();
    await expect(learner.getByRole("status")).toContainText(
      "verification instructions",
    );
    let messageID = "";
    await expect
      .poll(
        async () => {
          const r = await request.get("http://localhost:58025/api/v1/search", {
            params: { query: "to:" + address },
          });
          const data = await r.json();
          messageID = data.messages?.[0]?.ID ?? "";
          return messageID;
        },
        { timeout: 15000 },
      )
      .not.toBe("");
    const mail = await (
      await request.get("http://localhost:58025/api/v1/message/" + messageID)
    ).json();
    const link = String(mail.Text).match(
      /http:\/\/localhost:3000\/verify#[a-f0-9]+/,
    )?.[0];
    expect(link).toBeTruthy();
    await learner.goto(link!);
    await learner
      .getByRole("button", { name: "Verify email", exact: true })
      .click();
    await expect(learner.getByRole("status")).toContainText("Email verified");
    await page.getByRole("button", { name: "Refresh", exact: true }).click();
    await page.getByRole("button", { name: "People", exact: true }).click();
    await page
      .getByRole("button", { name: "Enrollment intake", exact: true })
      .click();
    const item = page
      .locator("article.queue-item")
      .filter({ hasText: address });
    await expect(item).toContainText("pending");
    await item
      .getByLabel("Course (for placement)", { exact: true })
      .selectOption({ label: "Foundations of Science" });
    await item
      .getByLabel("Start date (for placement)", { exact: true })
      .fill("2026-11-17");
    await item
      .getByLabel("Decision reason", { exact: true })
      .fill("Approved synthetic browser placement");
    page.once("dialog", (d) => d.accept());
    await item
      .getByRole("button", { name: "Apply decision", exact: true })
      .click();
    await expect(item).toContainText("placed");
    await learner.goto("http://localhost:3000/sign-in");
    await learner.getByLabel("Email address", { exact: true }).fill(address);
    await learner
      .getByLabel("Password", { exact: true })
      .fill("a browser registration phrase 3!");
    await learner.getByRole("button", { name: "Sign in", exact: true }).click();
    await expect(
      learner.getByRole("heading", { name: "My account", exact: true }),
    ).toBeVisible();
    await expect(
      learner
        .getByRole("navigation")
        .getByRole("button", { name: "Institution settings", exact: true }),
    ).toHaveCount(0);
  } finally {
    await context.close();
  }
});
