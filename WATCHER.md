Watcher integration (Linux only)

Overview
- vm_server now starts a fanotify-based watcher (when enabled) and records "consumers" — processes that access files under the requested scan paths.
- Consumers are upserted with FilePath, Comm, Exe, RUID, EUID and timestamps (first/last seen).

Flags
- `-watch-enable` (default: true): enable/disable watcher.
- `-watch-mount` (default: false): watch the whole mount for each provided path (more coverage, noisier). When false, a directory mark is used (non-recursive; child events only).
- `-watch-verbose` (default: false): verbose watcher logs.

API
- Start a scan: `POST /api/v1/scan` with body `{ "paths": ["/abs/dir"], "regexes": [], "threshold": 1 }`
- List consumers: `GET /api/v1/consumers?filePath=/abs/file/path`

Ubuntu VM quick start
1) Build vm_server (Linux):
   - `cd vm_server`
   - `go mod tidy`
   - `GOOS=linux GOARCH=$(uname -m | sed 's/x86_64/amd64/;s/aarch64/arm64/') go build -o eso-vm-server .`
2) Copy to VM and run as root (fanotify needs privileges):
   - `sudo ./eso-vm-server -port=1323 -watch-enable=true -watch-verbose=true`
3) Start a scan against a directory:
   - `curl -s -X POST http://localhost:1323/api/v1/scan -H 'content-type: application/json' -d '{"paths":["/home/ubuntu/watch"],"regexes":[],"threshold":1}'`
4) Generate activity under the path (new shell):
   - `echo hi >> /home/ubuntu/watch/file && cat /home/ubuntu/watch/file`
5) Query consumers:
   - `curl -s 'http://localhost:1323/api/v1/consumers?filePath=/home/ubuntu/watch/file' | jq`

Notes
- Directory marks are not recursive; events fire for direct children. Use `-watch-mount` for broad coverage.
- Running inside containers may require `--privileged`, `--pid=host`, and host bind mounts to see host activity.
- On non-Linux platforms, the watcher is disabled at build time; the server still runs without consumer tracking.

Permissions and capabilities
- Simplest: run as root (sudo). Fanotify and /proc reads work reliably.
- Non-root options:
  - Grant capabilities to the binary or via systemd:
    - `CAP_SYS_ADMIN` (required for mount marks; generally needed for fanotify in practice)
    - `CAP_DAC_READ_SEARCH` (improves ability to read `/proc/<pid>` for comm/exe/uid)
    - Example: `sudo setcap 'cap_sys_admin,cap_dac_read_search+ep' /usr/local/bin/eso-vm-server`
  - Ensure seccomp/AppArmor do not block fanotify syscalls (use unconfined profile if needed).
  - `/proc` may be mounted with `hidepid=2` on hardened systems; this blocks reading other PIDs. Options:
    - Run as root, or
    - Remount `/proc` with `hidepid=0` (security trade-off), or
    - Add `CAP_DAC_READ_SEARCH`.
  - Containers/K8s: use `--cap-add SYS_ADMIN`, `--pid=host`, and unconfined seccomp/AppArmor; or `--privileged` for PoC.

Startup checks
- On startup, the watcher emits warnings if:
  - `-watch-mount` is set but CAP_SYS_ADMIN is missing.
  - `/proc/<pid>/status` cannot be read for other processes (likely hidepid/permissions). It suggests fixes.
