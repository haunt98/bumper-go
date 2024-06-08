# bumper

[![Go](https://github.com/haunt98/bumper-go/actions/workflows/go.yml/badge.svg)](https://github.com/haunt98/bumper-go/actions/workflows/go.yml)
[![gitleaks](https://github.com/haunt98/bumper-go/actions/workflows/gitleaks.yml/badge.svg)](https://github.com/haunt98/bumper-go/actions/workflows/gitleaks.yml)
[![Latest Version](https://img.shields.io/github/v/tag/haunt98/bumper-go)](https://github.com/haunt98/bumper-go/tags)

Collection of bumping version scripts.

Because there is no 1 rule fit all.

So I made `bump00`, `bump01`, ... as time goes by.

## `bump00`

Install:

```sh
go install github.com/haunt98/bumper-go/cmd/bump00@latest
```

RC mode aka default mode:

```sh
bump00
```

- If latest tag is release, it will bump patch with `RC1`: `v1.2.3` ->
  `v1.2.4-RC1`
- If latest tag is RC, it will only bump RC: `v1.2.4-RC1` -> `v1.2.4-RC2` ->
  `v1.2.4-RC3`

Release mode:

```sh
bump00 --release
```

- If latest tag is release, it will bump patch: `v1.2.3` -> `v1.2.4`
- If latest tag is RC, it will only remove RC: `v1.2.4-RC1` -> `v1.2.4`

You can **mixed** RC and release mode.
