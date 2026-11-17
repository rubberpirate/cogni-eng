import { readFileSync, writeFileSync, mkdirSync, rmSync } from "node:fs";
import { randomBytes } from "node:crypto";
import { chromium, expect } from "@playwright/test";
const suffix = randomBytes(5).toString("hex"),
  network = `learning-check-${suffix}`,
  api = `learning-api-${suffix}`,
  web = `learning-web-${suffix}`;
const config = Object.fromEntries(
  readFileSync(".env", "utf8")
    .split("\n")
    .filter((x) => x && !x.startsWith("#"))
    .map((x) => {
      const i = x.indexOf("=");
      return [x.slice(0, i), x.slice(i + 1)];
    }),
);
if (config.APP_ENV !== "development")
  throw new Error(
    "Container qualification requires the synthetic development instance.",
  );
async function command(args: string[], capture = false) {
  const p = Bun.spawn(args, {
    stdout: capture ? "pipe" : "inherit",
    stderr: "inherit",
    env: { ...process.env, ...config },
  });
  const text = capture
    ? await new Response(p.stdout as ReadableStream).text()
    : "";
  if ((await p.exited) !== 0) throw new Error(`${args[0]} operation failed`);
  return text.trim();
}
await command(["go", "run", "./cmd/operator", "browser-fixture"]);
const credentials = readFileSync(".local/browser-access.txt", "utf8");
const email = credentials.match(/Email: (.+)/)![1],
  password = credentials.match(/Password: (.+)/)![1];
const services: Record<string, string> = {};
for (const service of ["database", "objects", "mail"]) {
  services[service] = await command(
    ["docker", "compose", "ps", "-q", service],
    true,
  );
  if (!services[service]) throw new Error(`Start local ${service} first.`);
}
mkdirSync(".local", { recursive: true, mode: 0o700 });
const envFile = `.local/container-${suffix}.env`;
writeFileSync(
  envFile,
  Object.entries({
    ...config,
    HTTP_ADDR: "0.0.0.0:8080",
    PUBLIC_ORIGIN: "http://localhost:13000",
    DATABASE_URL: `postgres://workspace:${config.DATABASE_PASSWORD}@database:5432/workspace?sslmode=disable`,
    OBJECT_ENDPOINT: "http://objects:8333",
    OBJECT_BUCKET: "private-workspace",
  })
    .map(([k, v]) => `${k}=${v}`)
    .join("\n"),
  { mode: 0o600 },
);
let created = false;
try {
  await command(["docker", "network", "create", network], true);
  created = true;
  for (const [alias, id] of Object.entries(services))
    await command([
      "docker",
      "network",
      "connect",
      "--alias",
      alias,
      network,
      id,
    ]);
  await command(
    [
      "docker",
      "run",
      "-d",
      "--name",
      api,
      "--network",
      network,
      "--network-alias",
      "gateway",
      "--env-file",
      envFile,
      "-p",
      "127.0.0.1:18080:8080",
      "learning-workspace-api:local",
    ],
    true,
  );
  await command(
    [
      "docker",
      "run",
      "-d",
      "--name",
      web,
      "--network",
      network,
      "-p",
      "127.0.0.1:13000:3000",
      "learning-workspace-web:local",
    ],
    true,
  );
  await expect
    .poll(
      async () => {
        try {
          return (await fetch("http://localhost:18080/health/ready")).status;
        } catch {
          return 0;
        }
      },
      { timeout: 30000 },
    )
    .toBe(200);
  await expect
    .poll(
      async () => {
        try {
          return (await fetch("http://localhost:13000/sign-in")).status;
        } catch {
          return 0;
        }
      },
      { timeout: 30000 },
    )
    .toBe(200);
  const browser = await chromium.launch({ headless: true });
  try {
    const page = await browser.newPage();
    await page.goto("http://localhost:13000/sign-in");
    await page.getByLabel("Email address", { exact: true }).fill(email);
    await page.getByLabel("Password", { exact: true }).fill(password);
    await page.getByRole("button", { name: "Sign in", exact: true }).click();
    await expect(
      page.getByRole("heading", { name: "People", exact: true }),
    ).toBeVisible();
    const results = await page.evaluate(async () => {
      const me = await (await fetch("/api/v1/me")).json();
      const c = document.createElement("canvas");
      c.width = 64;
      c.height = 64;
      const ctx = c.getContext("2d")!;
      ctx.fillStyle = "#276249";
      ctx.fillRect(0, 0, 64, 64);
      const body = await new Promise<Blob | null>((r) =>
        c.toBlob(r, "image/png"),
      );
      const put = await fetch("/api/v1/profile/avatar", {
        method: "PUT",
        headers: { "X-CSRF-Token": me.csrf, "Content-Type": "image/png" },
        body,
      });
      const read = await fetch("/api/v1/profile/avatar");
      const remove = await fetch("/api/v1/profile/avatar", {
        method: "DELETE",
        headers: { "X-CSRF-Token": me.csrf },
      });
      return {
        put: put.status,
        read: read.status,
        private: read.headers.get("Cache-Control"),
        remove: remove.status,
      };
    });
    expect(results).toEqual({
      put: 200,
      read: 200,
      private: "no-store",
      remove: 200,
    });
    await page.screenshot({
      path: ".local/screenshots/container-workspace.png",
      fullPage: true,
    });
  } finally {
    await browser.close();
  }
  console.log(
    "Container journey passed: browser → web proxy → Go → PostgreSQL and private S3.",
  );
} catch (error) {
  for (const name of [api, web]) {
    const p = Bun.spawn(["docker", "logs", "--tail", "30", name], {
      stdout: "inherit",
      stderr: "inherit",
    });
    await p.exited;
  }
  throw error;
} finally {
  for (const name of [web, api]) {
    const p = Bun.spawn(["docker", "rm", "-f", name], {
      stdout: "ignore",
      stderr: "ignore",
    });
    await p.exited;
  }
  if (created) {
    for (const id of Object.values(services)) {
      const p = Bun.spawn(["docker", "network", "disconnect", network, id], {
        stdout: "ignore",
        stderr: "ignore",
      });
      await p.exited;
    }
    await command(["docker", "network", "rm", network], true);
  }
  rmSync(envFile, { force: true });
}
