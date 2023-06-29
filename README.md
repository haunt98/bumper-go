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

Release mode:

```sh
bump00 --release
```

Only work for common, grpc pkg which don't need RC (version only have major, minor, patch).

Do not **mixed** RC and release mode.
They will fuck you up for good.
