const { test } = require("node:test");
const assert = require("node:assert/strict");
const fs = require("node:fs");
const os = require("node:os");
const path = require("node:path");
const { spawnSync } = require("node:child_process");

function fixture(t) {
  const root = fs.mkdtempSync(path.join(os.tmpdir(), "learning server "));
  t.after(() => fs.rmSync(root, { recursive: true, force: true }));
  for (const dir of ["scripts", "deploy", "tools", "home/.cloudflared"])
    fs.mkdirSync(path.join(root, dir), { recursive: true });
  fs.copyFileSync(
    path.join(__dirname, "server.sh"),
    path.join(root, "scripts/server.sh"),
  );
  const config = path.join(__dirname, "../deploy/ecosystem.config.cjs");
  fs.copyFileSync(config, path.join(root, "deploy/ecosystem.config.cjs"));
  const env = {
    PATH: `${root}/tools:${process.env.PATH}`,
    HOME: `${root}/home`,
    TEST_CALLS: `${root}/calls`,
  };
  function tool(name, body) {
    fs.writeFileSync(
      `${root}/tools/${name}`,
      `#!/bin/bash\nset -eu\n${body}\n`,
      { mode: 0o700 },
    );
  }
  tool(
    "go",
    [
      'if [[ "$*" == "version" ]]; then echo "go version go1.27.4 linux/amd64"; exit 0; fi',
      '[[ "${FAIL_BUILD:-}" != "${@: -1}" ]] || exit 23',
      'while [[ "$1" != -o ]]; do shift; done',
      "cat > \"$2\" <<'MOCK'",
      "#!/bin/bash",
      "printf '%s\\n' \"$DATABASE_URL|$OBJECT_PASSWORD|$*|${MIGRATION_DATABASE_URL-unset}\"",
      "MOCK",
      'chmod 700 "$2"',
    ].join("\n"),
  );
  tool(
    "pm2",
    `
    [[ -z "\${DATABASE_URL:-}" && -z "\${OBJECT_PASSWORD:-}" ]] || exit 24
    printf '%s\\n' "$*" >> "$TEST_CALLS"
  `,
  );
  tool("cloudflared", `printf '%s\\n' "$*" >> "$TEST_CALLS"`);
  tool("docker", `printf 'docker %s\\n' "$*" >> "$TEST_CALLS"`);
  tool("curl", `printf 'curl %s\\n' "$*" >> "$TEST_CALLS"`);
  const production =
    [
      "APP_ENV=production",
      "HTTP_ADDR=127.0.0.1:8080",
      "PUBLIC_ORIGIN=https://learn.example.com",
      `MFA_KEY=${"ab".repeat(32)}`,
      "DATABASE_URL=runtime-db",
      "MIGRATION_DATABASE_URL=migration-db",
      "OBJECT_ENDPOINT=https://s3.example.com",
      "OBJECT_BUCKET=private",
      "OBJECT_USER=user",
      "OBJECT_PASSWORD='secret $value & spaces'",
      "SMTP_ADDR=smtp.example.com:587",
      "SMTP_FROM=security@example.com",
    ].join("\n") + "\n";
  fs.writeFileSync(`${root}/.env.production`, production, { mode: 0o600 });
  function run(args, extra = {}) {
    return spawnSync("/bin/bash", ["scripts/server.sh", ...args], {
      cwd: root,
      env: { ...env, ...extra },
      encoding: "utf8",
    });
  }
  return { root, run, production };
}

test("failed builds keep the previous binaries and remove partial output", (t) => {
  const { root, run } = fixture(t);
  assert.equal(run(["build"]).status, 0);
  const previous = fs.readlinkSync(`${root}/.local/server/current`);
  const failure = run(["build"], { FAIL_BUILD: "./cmd/worker" });
  assert.equal(failure.status, 23, failure.stderr);
  assert.equal(fs.readlinkSync(`${root}/.local/server/current`), previous);
  assert.equal(fs.readdirSync(`${root}/.local/server/releases`).length, 1);
  assert.equal(run(["build"]).status, 0);
  assert.notEqual(fs.readlinkSync(`${root}/.local/server/current`), previous);
  assert.equal(fs.existsSync(`${root}/.local/server/${previous}/worker`), true);
});

test("migration owner is scoped to migrations and runtime reloads the private env", (t) => {
  const { root, run } = fixture(t);
  assert.equal(run(["build"]).status, 0);
  const migration = run(["migrate"]);
  assert.equal(migration.status, 0, migration.stderr);
  assert.match(
    migration.stdout,
    /^migration-db\|secret \$value & spaces\|migrate\|/,
  );
  const runtime = run(["run", "worker"]);
  assert.equal(runtime.status, 0, runtime.stderr);
  assert.match(runtime.stdout, /^runtime-db\|secret \$value & spaces\|\|unset/);
  const envFile = `${root}/.env.production`;
  fs.writeFileSync(
    envFile,
    fs.readFileSync(envFile, "utf8").replace("runtime-db", "updated-db"),
  );
  assert.match(run(["run", "gateway"]).stdout, /^updated-db\|/);
  assert.equal(run(["start"]).status, 0);
  assert.match(
    fs.readFileSync(`${root}/calls`, "utf8"),
    /--only learning-api,learning-worker/,
  );
});

test("invalid production origins and public bind addresses cannot start PM2", (t) => {
  const { root, run, production } = fixture(t);
  assert.equal(run(["build"]).status, 0);
  for (const invalid of [
    production.replace("https://learn.example.com", "http://learn.example.com"),
    production.replace("127.0.0.1:8080", "0.0.0.0:8080"),
  ]) {
    fs.writeFileSync(`${root}/.env.production`, invalid);
    assert.notEqual(run(["start"]).status, 0);
    assert.equal(fs.existsSync(`${root}/calls`), false);
  }
});

test("the tunnel starts separately and PM2 gives Go time to shut down", (t) => {
  const { root, run } = fixture(t);
  assert.notEqual(run(["tunnel"]).status, 0);
  fs.writeFileSync(`${root}/home/.cloudflared/config.yml`, "tunnel: test\n");
  assert.equal(run(["tunnel"]).status, 0);
  assert.equal(run(["run-tunnel"]).status, 0);
  const calls = fs.readFileSync(`${root}/calls`, "utf8");
  assert.match(calls, /--only learning-tunnel/);
  assert.ok(
    calls.includes(`tunnel --config ${root}/home/.cloudflared/config.yml run`),
  );
  const { apps } = require(`${root}/deploy/ecosystem.config.cjs`);
  assert.equal(apps.length, 3);
  for (const app of apps) {
    assert.equal(app.cwd, root);
    assert.equal(app.instances, 1);
    assert.equal(app.interpreter, "/bin/bash");
    assert.ok(app.kill_timeout > 15000);
  }
});

test("check command audits environment and restart-scratch runs full initialization sequence", (t) => {
  const { root, run } = fixture(t);
  const check = run(["check"]);
  assert.equal(check.status, 0, check.stderr);
  assert.match(check.stdout, /Preflight & Environment Diagnostic Check/);
  assert.match(check.stdout, /PUBLIC_ORIGIN: https:\/\/learn\.example\.com/);

  // create a dummy tls directory so restart-scratch proceeds
  fs.mkdirSync(`${root}/.local/server/tls`, { recursive: true });
  fs.writeFileSync(`${root}/.local/server/tls/ca.crt`, "ca");
  fs.writeFileSync(`${root}/.local/server/tls/server.crt`, "crt");

  const restart = run(["restart-scratch"]);
  assert.equal(restart.status, 0, restart.stderr);
  assert.match(restart.stdout, /Starting Cogni Backend Scratch \/ Restart Sequence/);
  const calls = fs.readFileSync(`${root}/calls`, "utf8");
  assert.match(calls, /docker compose --env-file .* -f compose\.server\.yaml up -d --wait/);
  assert.match(calls, /startOrRestart .* --only learning-api,learning-worker/);
  assert.match(calls, /curl .* http:\/\/127\.0\.0\.1:8080\/health\/ready/);
});
