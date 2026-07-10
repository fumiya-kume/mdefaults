# Testing Strategy

## Goal

The test suite protects the public CLI contract and every safety-sensitive error
path without invoking the host's macOS `defaults` command.

Coverage is split into three complementary layers:

1. Unit tests cover parsing, type validation, command argument construction,
   private and atomic filesystem behavior, and pull/push sequencing.
2. Docker E2E tests execute the compiled CLI against a stateful fake `defaults`
   executable with an isolated `HOME`, working directory, and `PATH`.
3. A macOS hosted-runner smoke test verifies the real `defaults` contract using
   a unique temporary domain that is always deleted during cleanup.

## Unit Tests

Run the complete race-enabled unit suite and enforce the coverage gate:

```sh
go test -race -shuffle=on -count=1 -covermode=atomic -coverprofile=coverage.out ./...
./scripts/check_coverage.sh coverage.out 95.0
```

All pure logic and injected adapters are expected to have full statement and
error-path coverage. The remaining entrypoint wrappers are exercised by E2E.

## Docker E2E

The Docker suite must be used for local black-box execution of the binary. It
cannot reach the host `defaults` command.

```sh
docker build -f test/e2e/Dockerfile -t mdefaults-e2e:local .
docker run --rm \
  --network none \
  --read-only \
  --tmpfs /tmp:rw,nosuid,nodev,size=64m \
  --cap-drop ALL \
  --security-opt no-new-privileges \
  mdefaults-e2e:local
```

The fake command records exact argument arrays as JSON, so spaces, empty values,
newlines, and shell metacharacters cannot be mistaken for argument boundaries.
The application binary is built with Go coverage instrumentation and prints its
E2E coverage report after the suite.

## Real macOS Smoke Test

The real-command test is intentionally guarded and runs only on an isolated
GitHub-hosted macOS runner:

```sh
go build -o "$RUNNER_TEMP/mdefaults" ./cmd/mdefaults
MDEFAULTS_RUN_REAL_DEFAULTS_SMOKE=1 \
MDEFAULTS_REAL_BINARY="$RUNNER_TEMP/mdefaults" \
go test -tags=realdefaults -count=1 -v ./test/macos
```

Never run this test against a developer workstation. The test uses a unique
domain, but the hosted runner is the required isolation boundary.

## Completion Criteria

- Unit coverage is at least 95% overall and safety-critical logic is fully
  covered.
- Unit tests pass with race detection and shuffled order.
- Docker E2E covers every public command, scalar type, prompt result, and error
  class, including migration of legacy config permissions to `0600`.
- The macOS smoke test confirms real type preservation.
- Build, vet, formatting, and `golangci-lint` remain clean.
