# Release and Upgrade Runbook

This repository builds `nstchain`.

## 1. Pre-release checklist

- [ ] Version and scope agreed; `CHANGELOG.md` updated.
- [ ] All intended commits are on `main` or `release/vX.Y.Z`.
- [ ] `make ci` passes.
- [ ] State-machine changes have a named upgrade handler.
- [ ] Store additions/removals are declared in `StoreUpgrades`.
- [ ] Migration tested on a restored mainnet-state copy.
- [ ] New binary starts, syncs, and serves RPC on a staging node.
- [ ] Rollback binary and data snapshot are available.

## 2. Build once

```bash
git checkout <release-commit>
make build-linux CMD=kava
sha256sum out/linux/nstchain
./out/linux/nstchain version --long
```

Record commit, version, Go version, build tags, and SHA-256. Do not rebuild on
individual production nodes.

## 3. Publish

```bash
git tag -s vX.Y.Z -m "vX.Y.Z"
git push origin vX.Y.Z
```

The release workflow builds the linux/amd64 artifact and publishes the binary
plus checksum. Verify the release before deployment.

## 4. Deploy

1. Snapshot every node's root volume (EBS or equivalent).
2. Back up `config/`, `genesis.json`, and `priv_validator_state.json`.
3. Upgrade the full/query node first.
4. Stop the old process, replace the binary, and start through systemd.
5. Verify chain ID, height, peers, `catching_up=false`, and `eth_chainId`.
6. Schedule the validator window. With fewer than three bonded validators, any
   validator restart is a chain halt; do not roll restart.
7. For a software upgrade, use a governance plan or an agreed halt height.
8. After the upgrade height, verify the handler log, module versions, and app
   hash agreement across nodes.

## 5. Verification commands

```bash
curl -s http://127.0.0.1:26657/status
curl -s -X POST -H 'Content-Type: application/json' \
  --data '{"jsonrpc":"2.0","method":"eth_chainId","params":[],"id":1}' \
  http://127.0.0.1:8545
nstchain query upgrade plan --home ~/.nstchain
nstchain query upgrade applied <upgrade-name> --home ~/.nstchain
nstchain query upgrade module_versions --home ~/.nstchain
```

Compare `app_hash` at the same height on every node.

## 6. Rollback

- Before the upgrade height: stop, restore the previous binary, restart.
- After state migration: restore the pre-upgrade data snapshot and old binary,
  then coordinate the network. Swapping only the binary is not a rollback.
- Record the incident, the failed height, and the exact artifacts used.
