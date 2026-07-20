# Security Policy

## Supported Versions

This is a fork of `killmonday/fscanx` focused on Windows 7-compatible
releases and a hardened CI/release pipeline. Security fixes are applied to
the latest `v*` release tag on the `master` branch.

| Version | Supported |
| ------- | --------- |
| latest `v*` | ✅ |
| older  | ❌ |

## Reporting a Vulnerability

This repository is a network security / intranet asset-survey tool. It is
intended only for authorized testing in environments you own or are
explicitly permitted to assess.

If you discover a vulnerability (e.g. RCE in a parsing path, credential
handling flaw, or CI supply-chain issue), please report it privately:

- Open a [GitHub Security Advisory](https://github.com/gandli/fscanx/security/advisories/new)
  (preferred — keeps the issue private until a fix is ready)
- Or email the maintainer via a GitHub issue marked **confidential** if the
  advisory form is unavailable.

Do **not** open a public issue for confirmed exploitable vulnerabilities.

## Supply-chain notes

- All GitHub Actions in `.github/workflows/*` are pinned to 40-char commit
  SHAs (not floating tags).
- The release pipeline uses Go 1.20 (last toolchain with Windows 7 runtime
  support) and UPX `v4.2.4`, both pinned.
- Dependabot is configured to **ignore major-version bumps**; they require
  manual review before merging.
