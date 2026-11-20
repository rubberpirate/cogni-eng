// Bun loads .env for configuration; pass only inference-specific environment to Ollama.
const endpoint = new URL(
  process.env.LOCAL_MODEL_URL ?? "http://127.0.0.1:11434",
);
if (
  endpoint.protocol !== "http:" ||
  !["localhost", "127.0.0.1", "[::1]"].includes(endpoint.hostname)
)
  throw new Error("LOCAL_MODEL_URL must be an HTTP loopback address.");
try {
  const response = await fetch(new URL("/api/version", endpoint), {
    signal: AbortSignal.timeout(2000),
  });
  if (response.ok) {
    console.log(`Ollama is already serving at ${endpoint.origin}.`);
    process.exit(0);
  }
} catch {}
const env: Record<string, string> = {};
for (const key of [
  "PATH",
  "HOME",
  "TMPDIR",
  "XDG_CACHE_HOME",
  "CUDA_VISIBLE_DEVICES",
  "NVIDIA_VISIBLE_DEVICES",
  "LD_LIBRARY_PATH",
])
  if (process.env[key]) env[key] = process.env[key]!;
Object.assign(env, {
  OLLAMA_HOST: endpoint.host,
  OLLAMA_NO_CLOUD: "1",
  OLLAMA_NUM_PARALLEL: "1",
});
const child = Bun.spawn(["ollama", "serve"], {
  env,
  stdout: "inherit",
  stderr: "inherit",
});
for (const signal of ["SIGINT", "SIGTERM"] as const)
  process.on(signal, () => child.kill(signal));
process.exit(await child.exited);
