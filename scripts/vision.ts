import { readFileSync, existsSync, writeFileSync, chmodSync } from "node:fs";
import { randomBytes } from "node:crypto";
if (!existsSync(".env"))
  throw new Error("Run bun run dev once before starting model inference.");
let text = readFileSync(".env", "utf8");
if (!/^VISION_API_KEY=.+$/m.test(text)) {
  text += `\nVISION_API_KEY=${randomBytes(32).toString("hex")}\nVISION_BASE_URL=http://127.0.0.1:58010/v1\n`;
  writeFileSync(".env", text, { mode: 0o600 });
  chmodSync(".env", 0o600);
}
const privateConfig = Object.fromEntries(
  text
    .split("\n")
    .filter((line) => line && !line.startsWith("#"))
    .map((line) => {
      const i = line.indexOf("=");
      return [line.slice(0, i), line.slice(i + 1)];
    }),
);
const args = process.argv.slice(2);
const runtimeEnv: Record<string, string> = {};
for (const key of [
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
]) {
  const value = process.env[key];
  if (value !== undefined) runtimeEnv[key] = value;
}
if (!existsSync("inference/.venv/pyvenv.cfg") || args.includes("--install")) {
  const installation = Bun.spawn(
    ["uv", "sync", "--project", "inference", "--frozen"],
    { env: runtimeEnv, stdout: "inherit", stderr: "inherit" },
  );
  if ((await installation.exited) !== 0)
    throw new Error(
      "Frozen inference environment installation failed. Existing sources are retained.",
    );
}
const runtime = Bun.spawn(
  [
    "uv",
    "run",
    "--project",
    "inference",
    "--frozen",
    "--no-sync",
    "python",
    "inference/server.py",
    ...args.filter((arg) => arg !== "--install"),
  ],
  {
    env: {
      ...runtimeEnv,
      VISION_API_KEY: privateConfig.VISION_API_KEY,
      HF_HOME: process.cwd() + "/.local/models",
      UV_LOCK_TIMEOUT: "1800",
      HF_HUB_DISABLE_TELEMETRY: "1",
      DO_NOT_TRACK: "1",
    },
    stdout: "inherit",
    stderr: "inherit",
  },
);
for (const signal of ["SIGINT", "SIGTERM"] as const)
  process.on(signal, () => runtime.kill());
process.exit(await runtime.exited);
