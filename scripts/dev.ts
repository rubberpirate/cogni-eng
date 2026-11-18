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
const commands = [
  [".local/bin/gateway"],
  [".local/bin/worker"],
  [
    "bun",
    "--no-env-file",
    "--bun",
    "apps/web/node_modules/next/dist/bin/next",
    "dev",
    "apps/web",
    "--hostname",
    "127.0.0.1",
  ],
];
if (config.VISION_AUTOSTART === "1")
  commands.push([
    "uv",
    "run",
    "--project",
    "inference",
    "--frozen",
    "--no-sync",
    "python",
    "inference/server.py",
  ]);
const children = commands.map((cmd) => {
  const childEnv = { ...env };
  if (cmd[0] !== ".local/bin/worker") {
    delete childEnv.GEMINI_API_KEY;
    delete childEnv.ANTHROPIC_API_KEY;
  }
  if (cmd[0] !== ".local/bin/worker" && !cmd.includes("inference/server.py"))
    delete childEnv.VISION_API_KEY;
  if (cmd.includes("inference/server.py"))
    for (const key of Object.keys(childEnv))
      if (
        ![
          "PATH",
          "HOME",
          "TMPDIR",
          "XDG_CACHE_HOME",
          "LD_LIBRARY_PATH",
          "CUDA_VISIBLE_DEVICES",
          "NVIDIA_VISIBLE_DEVICES",
          "NVIDIA_DRIVER_CAPABILITIES",
          "SSL_CERT_FILE",
          "SSL_CERT_DIR",
          "HTTPS_PROXY",
          "HTTP_PROXY",
          "NO_PROXY",
          "VISION_API_KEY",
        ].includes(key)
      )
        delete (childEnv as Record<string, string | undefined>)[key];
  if (cmd.includes("inference/server.py"))
    Object.assign(childEnv, {
      HF_HOME: process.cwd() + "/.local/models",
      HF_HUB_DISABLE_TELEMETRY: "1",
      DO_NOT_TRACK: "1",
    });
  const child = Bun.spawn(cmd, {
    env: childEnv,
    stdout: "inherit",
    stderr: "inherit",
  });
  if (cmd.includes("inference/server.py"))
    void child.exited.then((code) => {
      if (code !== 0)
        console.error(
          "Vision runtime stopped. Sources and jobs are retained; restart it with bun run vision.",
        );
    });
  return child;
});
console.log(
  "Workspace: http://localhost:3000 · Local mail catcher: http://localhost:58025",
);
for (const signal of ["SIGINT", "SIGTERM"] as const)
  process.on(signal, () => {
    children.forEach((p) => p.kill());
    process.exit(0);
  });
const result = await Promise.race(
  children
    .filter((_, i) => !commands[i].includes("inference/server.py"))
    .map((p) => p.exited),
);
children.forEach((p) => p.kill());
if (result !== 0)
  throw new Error("An application process stopped. Review its startup error.");
