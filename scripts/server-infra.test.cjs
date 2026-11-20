const { test } = require("node:test");
const assert = require("node:assert/strict");
const fs = require("node:fs");
const os = require("node:os");
const path = require("node:path");
const { spawnSync } = require("node:child_process");

test("server infrastructure init creates matching TLS/env and preserves existing secrets", (t) => {
  const root = fs.mkdtempSync(path.join(os.tmpdir(), "learning-infra-"));
  t.after(() => fs.rmSync(root, { recursive: true, force: true }));
  fs.mkdirSync(`${root}/scripts`);
  fs.copyFileSync(
    `${__dirname}/server-infra.sh`,
    `${root}/scripts/server-infra.sh`,
  );
  const run = (...args) =>
    spawnSync("bash", ["scripts/server-infra.sh", ...args], {
      cwd: root,
      encoding: "utf8",
    });
  assert.notEqual(run("init", "http://example.com").status, 0);
  assert.equal(fs.existsSync(`${root}/.env.production`), false);
  const init = run("init", "https://learn.example.com");
  assert.equal(init.status, 0, init.stderr);
  const original = fs.readFileSync(`${root}/.env.production`, "utf8");
  const env = Object.fromEntries(
    original
      .trim()
      .split("\n")
      .map((line) => {
        const split = line.indexOf("=");
        return [
          line.slice(0, split),
          line.slice(split + 1).replace(/^'(.*)'$/, "$1"),
        ];
      }),
  );
  assert.ok(
    env.DATABASE_URL.includes(`workspace:${env.APP_DATABASE_PASSWORD}@`),
  );
  assert.ok(
    env.MIGRATION_DATABASE_URL.includes(
      `workspace_owner:${env.DATABASE_PASSWORD}@`,
    ),
  );
  assert.notEqual(env.DATABASE_PASSWORD, env.APP_DATABASE_PASSWORD);
  assert.match(env.MFA_KEY, /^[0-9a-f]{64}$/);
  assert.equal(env.OBJECT_ENDPOINT, "https://127.0.0.1:59443");
  assert.equal(fs.statSync(`${root}/.env.production`).mode & 0o777, 0o600);
  const verify = spawnSync(
    "openssl",
    [
      "verify",
      "-CAfile",
      env.PGSSLROOTCERT,
      "-verify_ip",
      "127.0.0.1",
      `${env.SERVER_TLS_DIR}/server.crt`,
    ],
    { encoding: "utf8" },
  );
  assert.equal(verify.status, 0, verify.stderr);
  assert.ok(
    fs
      .readFileSync(env.SSL_CERT_FILE, "utf8")
      .includes(fs.readFileSync(env.PGSSLROOTCERT, "utf8")),
  );
  assert.notEqual(run("init", "https://other.example.com").status, 0);
  assert.equal(fs.readFileSync(`${root}/.env.production`, "utf8"), original);
  assert.equal(init.stdout.includes(env.DATABASE_PASSWORD), false);
});
