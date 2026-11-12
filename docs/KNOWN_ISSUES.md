# Known Issues

These are known, reproducible issues in the source baseline. They are not
silenced in CI; the smoke gate simply does not include the affected packages
until they are fixed.

## Test-suite rebrand debt

- `x/evmutil/...` and `x/earn/...` now use the `nst` prefix and the `unst`
  staking/savings denom; both package trees run in the CI gate.
- The shared test app keeps `ukava` as its default bond denom for unmodified
  suites. Tests that need the production denom pass an explicit staking
  genesis (see `x/earn/testutil`) or use `app.TestBondDenom`.
- The current CI gate runs `./app`, `./client/rest`, `./x/evmutil/...`, and
  `./x/earn/...`. Expand it as packages are repaired. Do not delete failing
  tests or mark them skipped to make the gate green.

## EVM fork mismatch

- `evm.params.chain_config` advertises `shanghai_block=0` and
  `cancun_block=0`, but the bundled `go-ethereum v1.10.26` does not implement
  PUSH0/MCOPY/TSTORE. `eth_call` with `0x5f00` returns
  `invalid opcode: PUSH0`.
- Short term: compile contracts with `evmVersion: "paris"` or `"london"`.
- Long term: upgrade the EVM/geth stack, or change the chain config so it does
  not claim unsupported forks.

## Validator availability

- Only two BONDED validators exist. Any validator restart stops the chain.
  Do not use rolling restarts for validator upgrades; schedule a coordinated
  maintenance window or add a third validator first.
