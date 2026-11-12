# Git Workflow and Repository Governance

This document defines the default workflow for the chain repository. It is
intentionally boring: one long-lived branch, short-lived change branches,
protected release tags, and a single build artifact that is promoted through
environments.

## 1. Repository roles

Keep the repository split by responsibility:

| Repository | Purpose | Visibility |
| --- | --- | --- |
| chain core | Cosmos SDK / CometBFT / EVM node source | private preferred |
| ops | runbooks, deploy scripts, monitoring config | private |
| website | public site and wallet assets | private or public |
| release artifacts | binaries, checksums, wallet-listing packages | attached to releases |

Do not use Git submodules for these repositories. Pin versions and artifact
checksums in the ops repository instead.

## 2. Branch model

Use trunk-based development. `main` is the only long-lived branch.

| Branch | Created from | Merged into | Purpose |
| --- | --- | --- | --- |
| `main` | - | - | always buildable, deployable candidate |
| `feat/<ticket>-<slug>` | `main` | `main` | feature work |
| `fix/<ticket>-<slug>` | `main` | `main` | bug fix |
| `chore/<slug>` | `main` | `main` | tooling, CI, docs |
| `release/vX.Y.Z` | `main` | `main` | release stabilization only |
| `hotfix/vX.Y.Z` | release tag | `main` and release branch | emergency production fix |

Rules:

- Never push directly to `main`.
- One logical change per pull request.
- Prefer squash merge; keep the PR title as the final commit subject.
- Delete the branch after merge.
- Rebase behind `main`; do not merge `main` into a feature branch repeatedly.
- A release branch accepts only release-blocking fixes, version bumps, and
  migration tests.
- Do not force-push `main`, release branches, or tags.

## 3. Pull request gate

Every PR must include:

- what changed and why;
- whether it is state-machine breaking;
- whether a new upgrade handler is required;
- commands run and their actual results;
- rollback path for production changes.

Required checks:

1. formatting (`gofmt`);
2. `go mod verify`;
3. daemon build;
4. unit smoke tests;
5. `go vet` on selected packages;
6. manual review by at least one maintainer.

The smoke suite is intentionally small while the rebrand/denom cleanup is in
progress. Expand it as packages become green; never replace a failing suite
with a permanently skipped test.

## 4. Tags and releases

- Use annotated, signed tags: `vMAJOR.MINOR.PATCH`.
- A tag must be created from a commit on `main` or a `release/*` branch.
- Build the release binary once in CI; upload the binary and its SHA-256.
- Never build a production binary independently on each node.
- Include the commit hash and build version in `version --long`.
- Keep release notes in `CHANGELOG.md` or generated release notes.
- A mainnet release requires two-person approval and a locked artifact.

## 5. Consensus and upgrade branches

A change is consensus/state-machine breaking if it changes module state,
store keys, denom metadata, EVM chain config, or upgrade migration logic.

For such changes:

1. open an issue/proposal describing the migration;
2. add a new named upgrade handler and any required `StoreUpgrades`;
3. test against a restored copy of mainnet state;
4. put the binary on the full node first;
5. schedule a coordinated validator window;
6. record the upgrade name, height, binary checksum, and operator approvals.

Do not mix an EVM/geth/Cosmos SDK major upgrade with a routine hotfix.

## 6. Secrets and access

- No private keys, seed phrases, API tokens, or production `.env` files in Git.
- CI uses short-lived GitHub tokens and repository secrets/environments.
- Mainnet deployment uses a protected GitHub environment with required reviewers.
- SSH keys for node access are per-person, scoped, and never shared.
- Rotate a key immediately if it is exposed; do not rewrite history as the only
  response.

## 7. Repository hygiene

- Ignore OS metadata (`._*`, `.DS_Store`), build outputs, local data, and logs.
- Keep `main` linear and free of generated binaries.
- Do not commit `out/`, `build/bin/`, node data directories, or private keys.
- Prefer `git bundle` or release assets over copying a dirty working tree between
  servers.
- Document every long-lived branch in this file.

## 8. Migration checklist for the current repositories

- [ ] Choose the canonical chain repository and grant maintainers admin access.
- [ ] Import the deployed source at the production commit; do not assume a
      different fork is the deployable baseline.
- [ ] Add `.gitignore`, `.gitattributes`, CI, release workflow, and this policy.
- [ ] Protect `main`; require PR, review, and green checks.
- [ ] Create the first signed release tag and publish binary + SHA-256.
- [ ] Move runbooks and deploy scripts into the ops repository.
- [ ] Archive duplicate or experimental repositories instead of leaving them
      ambiguous.

## 9. GitHub settings for repository admins

Apply these in the canonical repository:

- Branches: protect `main`; require a pull request, one approval, conversation
  resolution, linear history, and the `CI / Build and unit smoke tests` check.
- Branches: block force pushes and deletions on `main`.
- Tags: protect `v*`; only release maintainers may create or delete them.
- Merge: allow squash merge only; delete the branch after merge.
- Actions: keep the default `GITHUB_TOKEN` read-only; grant `contents: write`
  only to the release workflow.
- Environments: create `mainnet` with required reviewers and restrict it to
  release tags.
- Security: enable private vulnerability reporting / Security Advisories if the
  plan allows it.

