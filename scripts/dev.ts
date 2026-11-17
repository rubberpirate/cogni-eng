import { existsSync, mkdirSync, readFileSync, writeFileSync } from "node:fs";
import { randomBytes } from "node:crypto";

if (!existsSync(".env")) {
  const db = randomBytes(24).toString("hex");
  writeFileSync(
    ".env",
    `APP_ENV=development\nHTTP_ADDR=127.0.0.1:8080\nPUBLIC_ORIGIN=http://localhost:3000\nDATABASE_PASSWORD=${db}\nDATABASE_URL=postgres://workspace:${db}@localhost:55432/workspace?sslmode=disable\nMFA_KEY=${randomBytes(32).toString("hex")}\nOBJECT_USER=workspace-local\nOBJECT_PASSWORD=${randomBytes(24).toString("hex")}\nSMTP_ADDR=localhost:51025\nSMTP_FROM=security@workspace.invalid\n`,
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
const env = { ...process.env, ...config };
async function run(cmd: string[]) {
  const p = Bun.spawn(cmd, { env, stdout: "inherit", stderr: "inherit" });
  if ((await p.exited) !== 0) throw new Error(`${cmd[0]} failed`);
}
await run(["docker", "compose", "up", "-d", "--wait"]);
await run(["go", "run", "./cmd/operator", "migrate"]);
await run(["go", "run", "./cmd/operator", "demo"]);
const children = [
  ["go", "run", "./cmd/gateway"],
  ["go", "run", "./cmd/worker"],
  ["bun", "run", "web"],
].map((cmd) => Bun.spawn(cmd, { env, stdout: "inherit", stderr: "inherit" }));
console.log(
  "Workspace: http://localhost:3000 · Local mail catcher: http://localhost:58025",
);
for (const signal of ["SIGINT", "SIGTERM"] as const)
  process.on(signal, () => {
    children.forEach((p) => p.kill());
    process.exit(0);
  });
await Promise.all(children.map((p) => p.exited));
