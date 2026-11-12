# Known Issues

These are known, reproducible issues in the source baseline. They are not
silenced in CI; the smoke gate simply does not include the affected packages
until they are fixed.

## Test-suite rebrand debt

- `x/evmutil/keeper` and `x/evmutil/types` tests still use the old `kava1...`
  Bech32 prefix and the old `aist` denom. The node now configures the `nst`
  prefix and `anst` EVM denom, so those tests fail before reaching the behavior
  under test.
- `x/earn` tests still use `ukava` while the local test app configures `uist`;
  `go test ./x/earn/...` fails with `invalid deposit denom`.
- The current CI gate therefore runs `./app`, `./client/rest`, `./x/evmutil`,
  and `./x/evmutil/migrations/v2`. Expand it as packages are repaired. Do not
  delete failing tests or mark them skipped to make the gate green.

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
