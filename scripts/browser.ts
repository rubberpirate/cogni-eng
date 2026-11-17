try {
  const ready = await fetch("http://localhost:3000/sign-in", {
    signal: AbortSignal.timeout(3000),
  });
  if (!ready.ok) throw new Error();
} catch {
  throw new Error(
    "Start the local application with bun run dev before browser tests.",
  );
}
const fixture = Bun.spawn(["go", "run", "./cmd/operator", "browser-fixture"], {
  stdout: "inherit",
  stderr: "inherit",
  env: process.env,
});
if ((await fixture.exited) !== 0) process.exit(1);
const test = Bun.spawn(
  ["bunx", "playwright", "test", ...process.argv.slice(2)],
  {
    stdout: "inherit",
    stderr: "inherit",
    env: { ...process.env, E2E_ACCESS_FILE: ".local/browser-access.txt" },
  },
);
process.exit(await test.exited);
