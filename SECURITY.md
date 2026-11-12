# Security Policy

## Reporting a vulnerability

Do not open a public issue for a suspected vulnerability.

Report it privately to the repository maintainers. Prefer GitHub Security
Advisories if they are enabled for the repository; otherwise contact a
maintainer through the private channel already used for this project.

Include:

- affected commit or release;
- impact and attack preconditions;
- reproduction steps or a minimal proof of concept;
- whether the issue affects consensus, keys, RPC exposure, or funds.

Do not include private keys, seed phrases, production credentials, or customer
data in a report. Redact them.

## Supported versions

Only the latest tagged release and the current `main` branch receive security
fixes. Older tags are historical artifacts unless a maintainer explicitly
backports a fix.

## Node and RPC operators

- Keep validator keys off shared machines and out of Git.
- Do not expose debug RPC namespaces to the public internet.
- Restrict `8545`, `8546`, `26657`, and `9090` by security group or reverse
  proxy; public endpoints need rate limiting and monitoring.
- Treat zero-fee and high-block-gas configurations as an abuse surface.
