import { existsSync, mkdirSync, readFileSync, writeFileSync } from "node:fs";
import { randomBytes } from "node:crypto";

if (!existsSync(".env")) {
  const db = randomBytes(24).toString("hex");
  writeFileSync(
    ".env",
    `APP_ENV=development\nHTTP_ADDR=127.0.0.1:8080\nPUBLIC_ORIGIN=http://localhost:3000\nDATABASE_PASSWORD=${db}\nDATABASE_URL=postgres://workspace:${db}@localhost:55432/workspace?sslmode=disable\nMFA_KEY=${randomBytes(32).toString("hex")}\nOBJECT_USER=workspace-local\nOBJECT_PASSWORD=${randomBytes(24).toString("hex")}\nSMTP_ADDR=localhost:51025\nSMTP_FROM=security@workspace.invalid\nOBJECT_ENDPOINT=http://localhost:59000\nOBJECT_BUCKET=private-workspace\n`,
    { mode: 0o600 },
  );
}
const config = Object.fromEntries(
  readFileSync(".env", "utf8")
    .trim()
    .split("\n")
    .filter((x) => x && !x.startsWith("#"))
    .map((x) => {
      const i = x.indexOf("=");
      return [x.slice(0, i), x.slice(i + 1)];
    }),
);
const env = {
  ...process.env,
  ...config,
  OBJECT_ENDPOINT: config.OBJECT_ENDPOINT ?? "http://localhost:59000",
  OBJECT_BUCKET: config.OBJECT_BUCKET ?? "private-workspace",
};
async function run(cmd: string[]) {
  const p = Bun.spawn(cmd, { env, stdout: "inherit", stderr: "inherit" });
  if ((await p.exited) !== 0) throw new Error(`${cmd[0]} failed`);
}
await run(["docker", "compose", "up", "-d", "--wait"]);
mkdirSync(".local/bin", { recursive: true, mode: 0o700 });
await Promise.all(
  ["gateway", "worker", "operator"].map((name) =>
    run(["go", "build", "-o", `.local/bin/${name}`, `./cmd/${name}`]),
  ),
);
await run([".local/bin/operator", "migrate"]);
await run([".local/bin/operator", "demo"]);
const children = [
  [".local/bin/gateway"],
  [".local/bin/worker"],
  [
    "bun",
    "--bun",
    "apps/web/node_modules/next/dist/bin/next",
    "dev",
    "apps/web",
    "--hostname",
    "127.0.0.1",
  ],
].map((cmd) => Bun.spawn(cmd, { env, stdout: "inherit", stderr: "inherit" }));
console.log(
  "Workspace: http://localhost:3000 · Local mail catcher: http://localhost:58025",
);
for (const signal of ["SIGINT", "SIGTERM"] as const)
  process.on(signal, () => {
    children.forEach((p) => p.kill());
    process.exit(0);
  });
const result = await Promise.race(children.map((p) => p.exited));
children.forEach((p) => p.kill());
if (result !== 0)
  throw new Error("An application process stopped. Review its startup error.");
