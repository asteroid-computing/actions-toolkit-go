// Shared launcher for the Go GitHub Actions. Each action's index.js is one line: require("../_shim/shim.js")(__dirname).
//
// The action name is the directory name, so it can never disagree with the binary name.
//
// Two delivery modes, chosen per runner platform:
//   1. Binary mode — dist/<name>_<os>_<arch> exists (committed by `./walle build actions`): exec it. Zero compile cost in CI.
//   2. Source mode — no binary for this platform: `go build` cmd/github-actions/<name> from this repository into RUNNER_TEMP and exec that. Needs actions/setup-go earlier in the job; costs one compile of the action's closure (cached across runs by the Go build cache).
// `./walle check actions` fails CI when a committed binary does not match its source, so binary mode cannot silently run stale code.
const fs = require("node:fs");
const os = require("node:os");
const path = require("node:path");
const { spawnSync } = require("node:child_process");

const goos = { linux: "linux", darwin: "darwin", win32: "windows" }[os.platform()];
const goarch = { x64: "amd64", arm64: "arm64" }[os.arch()];

function fail(message) {
  console.log(`::error::${message}`);
  process.exit(1);
}

function exec(bin) {
  const result = spawnSync(bin, { stdio: "inherit" });
  if (result.error) fail(`failed to execute ${bin}: ${result.error.message}`);
  process.exit(result.status ?? 1);
}

module.exports = function launch(actionDir) {
  const name = path.basename(actionDir);
  if (!goos || !goarch) fail(`unsupported runner platform ${os.platform()}/${os.arch()}`);

  const ext = goos === "windows" ? ".exe" : "";
  const bin = path.join(actionDir, "dist", `${name}_${goos}_${goarch}${ext}`);
  if (fs.existsSync(bin)) exec(bin);

  // Source mode. The module root is three levels up: <root>/.github/actions/<name>.
  const root = path.resolve(actionDir, "..", "..", "..");
  const out = path.join(process.env.RUNNER_TEMP || os.tmpdir(), `${name}${ext}`);
  console.log(`No committed binary for ${goos}/${goarch}; building ${name} from source.`);
  const build = spawnSync("go", ["build", "-trimpath", "-o", out, `./cmd/github-actions/${name}`], { cwd: root, stdio: "inherit" });
  if (build.error) fail(`go is not on PATH (${build.error.message}); add actions/setup-go before this step, or commit binaries with ./walle build actions`);
  if (build.status !== 0) fail(`building ${name} from source failed`);
  exec(out);
};
