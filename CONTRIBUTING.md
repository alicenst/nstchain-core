# Contributing

Read `docs/GIT_GOVERNANCE.md` before opening a pull request.

## Local checks

```bash
go mod verify
gofmt -l $(git ls-files '*.go')
go build ./cmd/kava
go test -count=1 ./app ./client/rest ./x/evmutil ./x/evmutil/migrations/v2
go vet ./app ./client/rest
```

`make ci` runs the same gate.

The full `go test ./...` suite still has known denom-rebrand failures. Do not
silence them with build tags or skipped tests; fix the affected package and
expand the smoke suite when it is green.

## Branch and commit

- Branch from `main`: `feat/<ticket>-<slug>`, `fix/<ticket>-<slug>`, or
  `chore/<slug>`.
- Keep one logical change per PR.
- Use Conventional Commits, for example `fix(rest): remove unused import`.
- Do not commit private keys, seed phrases, tokens, node data, or logs.
- Do not bump dependencies and change consensus logic in the same PR.

## Pull requests

Use the repository PR template. State explicitly whether the change is
state-machine breaking and whether it needs a new upgrade handler. Fill in the
validation section with commands actually run.
