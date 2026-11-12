# Source Baseline

The initial commit of this repository was imported from the NSTChain production
source tree at:

- commit: `53176d9a0f8211efe9f4a18cc0ace7a7bab45037`
- date: 2025-10-31
- subject: `change aist to anst`
- source: `nst-1:~/workspace/istchain`

The original repository history and its dirty working tree remain available on
`nst-1`. This repository starts a clean history on purpose:

- removed the tracked E2E mnemonic (`tests/e2e/.env`); use
  `tests/e2e/.env.example` and keep the real value local;
- removed the tracked macOS IDE directory `.idea/`;
- removed the tracked `build/bin/protoc` macOS binary;
- added repository governance, CI, release automation, and hygiene rules.

The binary built from the source baseline reports:

```text
name: nstchain
commit: 53176d9a0f8211efe9f4a18cc0ace7a7bab45037
cosmos_sdk_version: v0.47.15
go: go1.23.4 linux/amd64
```

Any future release must be built from a tag on this repository, not from the
dirty production tree.
