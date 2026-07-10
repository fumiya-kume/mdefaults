# Real macOS `defaults` smoke test

This smoke test is intentionally excluded from ordinary test runs. It requires
both the `realdefaults` build tag and an explicit environment guard because it
executes the real `/usr/bin/defaults` command.

The test uses an isolated `HOME`, creates a cryptographically unique defaults
domain, and registers cleanup before seeding any values. It verifies that
string, integer, boolean, float, and date values retain their values and native
types through an actual `mdefaults pull` followed by `mdefaults push`. It also
confirms that a real array value is rejected without rewriting the config file.

Build the binary under test, then opt in explicitly:

```sh
go build -o /tmp/mdefaults-real ./cmd/mdefaults
MDEFAULTS_RUN_REAL_DEFAULTS_SMOKE=1 MDEFAULTS_REAL_BINARY=/tmp/mdefaults-real go test -tags=realdefaults -count=1 -v ./test/macos
```
