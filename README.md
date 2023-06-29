# bumper

Collection of bumping version scripts.

Because there is no 1 rule fit all.

So I made `bump00`, `bump01`, ... as time goes by.

## `bump00`

RC mode:

```sh
bump00 --rc
```

Only work for services that need RC for testing.
When deploy production, need to create release from GitHub/GitLab.

- If latest version don't have RC, it will bump patch version with `RC1`: `v1.2.3` -> `v1.2.4-RC1`
- If latest version already have RC, it will only bump RC version: `v1.2.4-RC1` -> `v1.2.4-RC2` -> `v1.2.4-RC3`

Release mode:

```sh
bump00 --release
```

Only work for common, grpc pkg which don't need RC (version only have major, minor, patch).

- Only bump patch: `v1.2.3` -> `v1.2.4` -> `v1.2.5`

Do not **mixed** RC and release mode.
They will fuck you up for good.
