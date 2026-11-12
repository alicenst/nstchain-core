#!/usr/bin/env bash
# Guard against re-introducing the upstream Kava module path or command.
set -euo pipefail

status=0

# Go sources must not import the old upstream module path.
if git grep -n -E 'github\.com/(kava-labs|Kava-Labs)/kava([^A-Za-z0-9_-]|$)' -- '*.go' >/tmp/dekava-self.txt 2>/dev/null; then
	echo "FAIL: upstream Kava module path found in Go sources"
	cat /tmp/dekava-self.txt
	status=1
else
	echo "ok: no upstream Kava module path in Go sources"
fi

# The binary entry point must stay on cmd/nstchain.
if [ -d cmd/kava ]; then
	echo "FAIL: cmd/kava exists; use cmd/nstchain"
	status=1
else
	echo "ok: cmd/nstchain is the only command entry point"
fi

# Kava-Labs modules are allowed only for the pinned forks listed in NOTICE.
# Update this list together with go.mod and NOTICE when a fork is replaced.
allowed=$(
	cat <<'EOF'
github.com/Kava-Labs/opendb
github.com/kava-labs/cometbft
github.com/kava-labs/cosmos-sdk
github.com/kava-labs/ethermint
github.com/kava-labs/iavl
EOF
)

actual=$(
	{
		go list -m all 2>/dev/null | awk '{print $1}'
		go list -m all 2>/dev/null | awk -F'=>' '/=>/ {print $2}'
	} | awk '{print $1}' | grep -i 'kava-labs' | grep -v '^github.com/kava-labs/kava$' | sort -u
)

unexpected=$(comm -13 <(printf '%s\n' "$allowed" | sort -u) <(printf '%s\n' "$actual" | sort -u))
if [ -n "$unexpected" ]; then
	echo "FAIL: unexpected Kava-Labs dependency (add it to the allowlist and NOTICE):"
	printf '%s\n' "$unexpected"
	status=1
else
	echo "ok: Kava-Labs dependencies are in the allowlist"
fi

exit "$status"
