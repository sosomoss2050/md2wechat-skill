# md2wechat OpenClaw runtime contract

This OpenClaw skill package is intentionally thin:

- it declares the OpenClaw metadata needed for discovery and install gating
- it documents the external credential surface explicitly
- it delegates execution to an already-installed `md2wechat` runtime

Runtime lookup order:

1. `~/.openclaw/tools/md2wechat/md2wechat`
2. `~/.openclaw/tools/md2wechat/bin/md2wechat`
3. `md2wechat` on `PATH`

Runtime behavior:

- no network download at execution time
- no cache-based bootstrapper
- no fallback to a GitHub release fetch from `run.sh`

When the runtime is missing, the wrapper exits with a clear install hint instead of trying to self-provision.
