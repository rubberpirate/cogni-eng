const path = require("node:path");

const root = path.resolve(__dirname, "..");
const app = (name, args) => ({
  name,
  cwd: root,
  script: path.join(root, "scripts/server.sh"),
  args,
  interpreter: "/bin/bash",
  exec_mode: "fork",
  instances: 1,
  autorestart: true,
  watch: false,
  restart_delay: 5000,
  // Go's worker allows 15 seconds for queue shutdown; PM2 sends SIGINT first.
  kill_timeout: 30000,
  time: true,
});

module.exports = {
  apps: [
    app("learning-api", ["run", "gateway"]),
    app("learning-worker", ["run", "worker"]),
    app("learning-tunnel", ["run-tunnel"]),
  ],
};
