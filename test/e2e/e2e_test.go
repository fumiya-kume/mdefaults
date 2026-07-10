//go:build e2e

package e2e_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"runtime"
	"strconv"
	"strings"
	"testing"
)

const (
	binaryEnv       = "MDEFAULTS_E2E_BINARY"
	containerEnv    = "MDEFAULTS_E2E_CONTAINER"
	coverageDirEnv  = "MDEFAULTS_E2E_COVERAGE_DIR"
	fakeDefaultsEnv = "MDEFAULTS_E2E_FAKE_DEFAULTS"
	fixtureEnv      = "MDEFAULTS_E2E_FIXTURE"
	invocationEnv   = "MDEFAULTS_E2E_INVOCATIONS"
)

type fakeFixture struct {
	Entries  []fakeEntry   `json:"entries,omitempty"`
	Failures []fakeFailure `json:"failures,omitempty"`
}

type fakeEntry struct {
	Domain     string `json:"domain"`
	Key        string `json:"key"`
	ReadOutput string `json:"read_output"`
	TypeOutput string `json:"type_output"`
}

type fakeFailure struct {
	Operation string `json:"operation"`
	Domain    string `json:"domain"`
	Key       string `json:"key"`
	ExitCode  int    `json:"exit_code"`
	Stderr    string `json:"stderr,omitempty"`
}

type fakeInvocation struct {
	Args []string `json:"args"`
}

type sandbox struct {
	root           string
	home           string
	work           string
	fakeBin        string
	configPath     string
	fixturePath    string
	invocationPath string
	binaryPath     string
}

type commandResult struct {
	exitCode int
	stdout   string
	stderr   string
}

func TestMain(m *testing.M) {
	if os.Getenv(fakeDefaultsEnv) == "1" {
		os.Exit(runFakeDefaults())
	}

	if runtime.GOOS != "linux" || os.Getenv(containerEnv) != "1" {
		fmt.Fprintln(os.Stderr, "mdefaults E2E tests must run in the dedicated Linux container")
		os.Exit(2)
	}

	coverageDir := os.Getenv(coverageDirEnv)
	if coverageDir == "" {
		fmt.Fprintln(os.Stderr, "missing E2E coverage directory")
		os.Exit(2)
	}
	for _, path := range []string{coverageDir, os.Getenv("HOME"), os.Getenv("GOCACHE")} {
		if path == "" {
			continue
		}
		if err := os.MkdirAll(path, 0o700); err != nil {
			fmt.Fprintf(os.Stderr, "create E2E runtime directory %s: %v\n", path, err)
			os.Exit(2)
		}
	}

	exitCode := m.Run()
	if err := reportCoverage(coverageDir); err != nil {
		fmt.Fprintf(os.Stderr, "report E2E coverage: %v\n", err)
		if exitCode == 0 {
			exitCode = 1
		}
	}
	os.Exit(exitCode)
}

func TestVersionAndHelpPaths(t *testing.T) {
	t.Run("long version", func(t *testing.T) {
		s := newSandbox(t, fakeFixture{})
		result := s.run(t, "", "--version")

		requireExitCode(t, result, 0)
		requireContains(t, result.stdout, "Version: e2e-version")
		requireContains(t, result.stdout, "Architecture: e2e-architecture")
		s.requireInvocations(t)
		requirePathNotExist(t, s.configPath)
	})

	t.Run("short version", func(t *testing.T) {
		s := newSandbox(t, fakeFixture{})
		result := s.run(t, "", "-v")

		requireExitCode(t, result, 0)
		requireContains(t, result.stdout, "Version: e2e-version")
		requireContains(t, result.stdout, "Architecture: e2e-architecture")
		s.requireInvocations(t)
		requirePathNotExist(t, s.configPath)
	})

	t.Run("no arguments", func(t *testing.T) {
		s := newSandbox(t, fakeFixture{})
		result := s.run(t, "")

		requireExitCode(t, result, 0)
		requireContains(t, result.stdout+result.stderr, "Usage: mdefaults")
		requireContains(t, result.stdout+result.stderr, "pull")
		requireContains(t, result.stdout+result.stderr, "push")
		s.requireInvocations(t)
		requirePathNotExist(t, s.configPath)
	})

	t.Run("help flag", func(t *testing.T) {
		s := newSandbox(t, fakeFixture{})
		result := s.run(t, "", "--help")

		requireExitCode(t, result, 0)
		combined := result.stdout + result.stderr
		requireContainsFold(t, combined, "usage")
		requireContains(t, combined, "pull")
		requireContains(t, combined, "push")
		s.requireInvocations(t)
		requirePathNotExist(t, s.configPath)
	})
}

func TestConfigCommand(t *testing.T) {
	s := newSandbox(t, fakeFixture{})
	s.writeConfig(t, configLine("com.example.app", "greeting", "hello world", "string"))
	if err := os.Chmod(s.configPath, 0o644); err != nil {
		t.Fatalf("set legacy config mode: %v", err)
	}

	result := s.run(t, "", "config")

	requireExitCode(t, result, 0)
	requireContains(t, result.stdout, "com.example.app")
	requireContains(t, result.stdout, "greeting")
	requireContains(t, result.stdout, "hello world")
	info, err := os.Stat(s.configPath)
	if err != nil {
		t.Fatalf("stat migrated config: %v", err)
	}
	if actual := info.Mode().Perm(); actual != 0o600 {
		t.Fatalf("expected migrated config permissions, got %o", actual)
	}
	s.requireInvocations(t)
}

func TestFlagsBeforeCommand(t *testing.T) {
	s := newSandbox(t, fakeFixture{})

	result := s.run(t, "", "--verbose", "push")

	requireExitCode(t, result, 0)
	requireContains(t, result.stdout, "Success: Configurations pushed successfully")
	requireContains(t, s.applicationLog(t), "Verbose mode enabled")
	s.requireInvocations(t)
}

func TestPushScalarArguments(t *testing.T) {
	s := newSandbox(t, fakeFixture{})
	configContent := strings.Join([]string{
		configLine("com.example.app", "plain", "hello world", "string"),
		configLine("com.example.app", "integer", "42", "integer"),
		configLine("com.example.app", "boolean", "true", "boolean"),
		configLine("com.example.app", "float", "3.5", "float"),
		configLine("com.example.app", "date", "2026-07-11 01:23:45 +0000", "date"),
		configLine("com.example.app", "empty", "", "string"),
		configLine("com.example.app", "escaped", "line one\nline \"two\"\t$(not a shell)", "string"),
	}, "")
	s.writeConfig(t, configContent)

	result := s.run(t, "", "push")

	requireExitCode(t, result, 0)
	requireContains(t, result.stdout, "Success: Configurations pushed successfully")
	if actual := s.readConfig(t); actual != configContent {
		t.Fatalf("push changed the config file:\nwant %q\n got %q", configContent, actual)
	}
	s.requireInvocations(t,
		[]string{"write", "com.example.app", "plain", "hello world"},
		[]string{"write", "com.example.app", "integer", "-int", "42"},
		[]string{"write", "com.example.app", "boolean", "-bool", "true"},
		[]string{"write", "com.example.app", "float", "-float", "3.5"},
		[]string{"write", "com.example.app", "date", "-date", "2026-07-11 01:23:45 +0000"},
		[]string{"write", "com.example.app", "empty", ""},
		[]string{"write", "com.example.app", "escaped", "line one\nline \"two\"\t$(not a shell)"},
	)
}

func TestBooleanPullPushRoundTrip(t *testing.T) {
	testCases := []struct {
		name       string
		readOutput string
		writeValue string
	}{
		{name: "true", readOutput: "1\n", writeValue: "true"},
		{name: "false", readOutput: "0\n", writeValue: "false"},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			s := newSandbox(t, fakeFixture{Entries: []fakeEntry{{
				Domain:     "com.example.app",
				Key:        "enabled",
				ReadOutput: testCase.readOutput,
				TypeOutput: "Type is boolean\n",
			}}})
			s.writeConfig(t, "com.example.app enabled\n")

			pullResult := s.run(t, "", "pull", "-y")
			requireExitCode(t, pullResult, 0)
			pushResult := s.run(t, "", "push")
			requireExitCode(t, pushResult, 0)

			s.requireInvocations(t,
				[]string{"read", "com.example.app", "enabled"},
				[]string{"read-type", "com.example.app", "enabled"},
				[]string{"write", "com.example.app", "enabled", "-bool", testCase.writeValue},
			)
		})
	}
}

func TestPushErrorsAndPrevalidation(t *testing.T) {
	t.Run("missing config is created", func(t *testing.T) {
		s := newSandbox(t, fakeFixture{})
		result := s.run(t, "", "push")

		requireExitCode(t, result, 0)
		if actual := s.readConfig(t); actual != "" {
			t.Fatalf("expected an empty config, got %q", actual)
		}
		info, err := os.Stat(s.configPath)
		if err != nil {
			t.Fatalf("stat created config: %v", err)
		}
		if actual := info.Mode().Perm(); actual != 0o600 {
			t.Fatalf("expected private config permissions, got %o", actual)
		}
		s.requireInvocations(t)
	})

	t.Run("unsupported type prevents every write", func(t *testing.T) {
		s := newSandbox(t, fakeFixture{})
		s.writeConfig(t, configLine("com.example.app", "valid", "first", "string")+
			configLine("com.example.app", "unsupported", "(one, two)", "array"))

		result := s.run(t, "", "push")

		requireExitCode(t, result, 1)
		requireContains(t, result.stdout+result.stderr, "Error: Failed to push configurations")
		requireNotContains(t, result.stdout+result.stderr, "Success:")
		s.requireInvocations(t)
	})

	t.Run("missing value prevents every write", func(t *testing.T) {
		s := newSandbox(t, fakeFixture{})
		s.writeConfig(t, configLine("com.example.app", "valid", "first", "string")+
			"com.example.app missing\n")

		result := s.run(t, "", "push")

		requireExitCode(t, result, 1)
		requireNotContains(t, result.stdout+result.stderr, "Success:")
		s.requireInvocations(t)
	})

	t.Run("malformed config does not invoke defaults", func(t *testing.T) {
		s := newSandbox(t, fakeFixture{})
		s.writeConfig(t, "com.example.app malformed \"unterminated\n")

		result := s.run(t, "", "push")

		requireExitCode(t, result, 1)
		requireContains(t, s.applicationLog(t), "parse config line 1")
		s.requireInvocations(t)
	})

	t.Run("write failure stops later writes", func(t *testing.T) {
		s := newSandbox(t, fakeFixture{Failures: []fakeFailure{{
			Operation: "write",
			Domain:    "com.example.app",
			Key:       "second",
			ExitCode:  47,
			Stderr:    "injected write failure",
		}}})
		s.writeConfig(t, configLine("com.example.app", "first", "one", "string")+
			configLine("com.example.app", "second", "two", "integer")+
			configLine("com.example.app", "third", "three", "string"))

		result := s.run(t, "", "push")

		requireExitCode(t, result, 1)
		requireContains(t, result.stdout+result.stderr, "Error: Failed to push configurations")
		requireNotContains(t, result.stdout+result.stderr, "Success:")
		s.requireInvocations(t,
			[]string{"write", "com.example.app", "first", "one"},
			[]string{"write", "com.example.app", "second", "-int", "two"},
		)
	})

	t.Run("config creation failure", func(t *testing.T) {
		s := newSandbox(t, fakeFixture{})
		s.home = filepath.Join(s.root, "missing-parent", "home")

		result := s.run(t, "", "push")

		requireExitCode(t, result, 1)
		requireContains(t, s.applicationLog(t), "Failed to create config file")
		s.requireInvocations(t)
	})
}

func TestPullScalarRewrite(t *testing.T) {
	entries := []fakeEntry{
		{Domain: "com.example.app", Key: "plain", ReadOutput: "hello world\n", TypeOutput: "Type is string\n"},
		{Domain: "com.example.app", Key: "integer", ReadOutput: "42\n", TypeOutput: "Type is integer\n"},
		{Domain: "com.example.app", Key: "boolean", ReadOutput: "1\n", TypeOutput: "Type is boolean\n"},
		{Domain: "com.example.app", Key: "float", ReadOutput: "3.5\n", TypeOutput: "Type is real\n"},
		{Domain: "com.example.app", Key: "date", ReadOutput: "2026-07-11 01:23:45 +0000\n", TypeOutput: "Type is date\n"},
		{Domain: "com.example.app", Key: "escaped", ReadOutput: "line one\nline \"two\"\t\n", TypeOutput: "Type is string\n"},
	}
	s := newSandbox(t, fakeFixture{Entries: entries})
	s.writeConfig(t, strings.Join([]string{
		"com.example.app plain\n",
		"com.example.app integer\n",
		"com.example.app boolean\n",
		"com.example.app float\n",
		"com.example.app date\n",
		"com.example.app escaped\n",
	}, ""))

	result := s.run(t, "", "--verbose", "pull", "-y")

	requireExitCode(t, result, 0)
	requireContains(t, result.stdout, "Success: Configurations pulled successfully")
	requireContains(t, s.applicationLog(t), "Verbose mode enabled")
	expectedConfig := strings.Join([]string{
		configLine("com.example.app", "plain", "hello world", "string"),
		configLine("com.example.app", "integer", "42", "integer"),
		configLine("com.example.app", "boolean", "1", "boolean"),
		configLine("com.example.app", "float", "3.5", "float"),
		configLine("com.example.app", "date", "2026-07-11 01:23:45 +0000", "date"),
		configLine("com.example.app", "escaped", "line one\nline \"two\"\t", "string"),
	}, "")
	if actual := s.readConfig(t); actual != expectedConfig {
		t.Fatalf("unexpected pulled config:\nwant %q\n got %q", expectedConfig, actual)
	}
	expectedInvocations := make([][]string, 0, len(entries)*2)
	for _, entry := range entries {
		expectedInvocations = append(expectedInvocations,
			[]string{"read", entry.Domain, entry.Key},
			[]string{"read-type", entry.Domain, entry.Key},
		)
	}
	s.requireInvocations(t, expectedInvocations...)
}

func TestPullPromptPaths(t *testing.T) {
	fixture := fakeFixture{Entries: []fakeEntry{{
		Domain:     "com.example.app",
		Key:        "greeting",
		ReadOutput: "new value\n",
		TypeOutput: "Type is string\n",
	}}}
	original := configLine("com.example.app", "greeting", "old value", "string")
	updated := configLine("com.example.app", "greeting", "new value", "string")
	expectedInvocations := [][]string{
		{"read", "com.example.app", "greeting"},
		{"read-type", "com.example.app", "greeting"},
	}

	t.Run("yes updates", func(t *testing.T) {
		s := newSandbox(t, fixture)
		s.writeConfig(t, original)

		result := s.run(t, "yes\n", "pull")

		requireExitCode(t, result, 0)
		requireContains(t, result.stdout, "Success: Configurations pulled successfully")
		if actual := s.readConfig(t); actual != updated {
			t.Fatalf("expected updated config %q, got %q", updated, actual)
		}
		s.requireInvocations(t, expectedInvocations...)
	})

	t.Run("no preserves", func(t *testing.T) {
		s := newSandbox(t, fixture)
		s.writeConfig(t, original)

		result := s.run(t, "no\n", "pull")

		requireExitCode(t, result, 0)
		requireContains(t, result.stdout, "Operation cancelled")
		requireNotContains(t, result.stdout+result.stderr, "Success:")
		if actual := s.readConfig(t); actual != original {
			t.Fatalf("cancelled pull changed config: want %q, got %q", original, actual)
		}
		s.requireInvocations(t, expectedInvocations...)
	})

	t.Run("EOF preserves", func(t *testing.T) {
		s := newSandbox(t, fixture)
		s.writeConfig(t, original)

		result := s.run(t, "", "pull")

		requireExitCode(t, result, 1)
		requireNotContains(t, result.stdout+result.stderr, "Success:")
		if actual := s.readConfig(t); actual != original {
			t.Fatalf("failed prompt changed config: want %q, got %q", original, actual)
		}
		s.requireInvocations(t, expectedInvocations...)
	})
}

func TestPullCommandErrorsPreserveConfig(t *testing.T) {
	original := "com.example.app greeting\n"
	testCases := []struct {
		name        string
		fixture     fakeFixture
		invocations [][]string
	}{
		{
			name: "read failure",
			fixture: fakeFixture{Failures: []fakeFailure{{
				Operation: "read", Domain: "com.example.app", Key: "greeting", ExitCode: 41, Stderr: "read failed",
			}}},
			invocations: [][]string{{"read", "com.example.app", "greeting"}},
		},
		{
			name: "read type failure",
			fixture: fakeFixture{
				Entries: []fakeEntry{{
					Domain: "com.example.app", Key: "greeting", ReadOutput: "hello\n", TypeOutput: "Type is string\n",
				}},
				Failures: []fakeFailure{{
					Operation: "read-type", Domain: "com.example.app", Key: "greeting", ExitCode: 42, Stderr: "type failed",
				}},
			},
			invocations: [][]string{
				{"read", "com.example.app", "greeting"},
				{"read-type", "com.example.app", "greeting"},
			},
		},
		{
			name: "malformed type output",
			fixture: fakeFixture{Entries: []fakeEntry{{
				Domain: "com.example.app", Key: "greeting", ReadOutput: "hello\n", TypeOutput: "string\n",
			}}},
			invocations: [][]string{
				{"read", "com.example.app", "greeting"},
				{"read-type", "com.example.app", "greeting"},
			},
		},
		{
			name: "unsupported type",
			fixture: fakeFixture{Entries: []fakeEntry{{
				Domain: "com.example.app", Key: "greeting", ReadOutput: "(one, two)\n", TypeOutput: "Type is array\n",
			}}},
			invocations: [][]string{
				{"read", "com.example.app", "greeting"},
				{"read-type", "com.example.app", "greeting"},
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			s := newSandbox(t, testCase.fixture)
			s.writeConfig(t, original)

			result := s.run(t, "", "pull", "-y")

			requireExitCode(t, result, 1)
			requireContains(t, result.stdout+result.stderr, "Error: Failed to pull configurations")
			requireNotContains(t, result.stdout+result.stderr, "Success:")
			if actual := s.readConfig(t); actual != original {
				t.Fatalf("failed pull changed config: want %q, got %q", original, actual)
			}
			s.requireInvocations(t, testCase.invocations...)
		})
	}
}

func TestPullLaterFailureDoesNotPartiallyRewrite(t *testing.T) {
	original := "com.example.app first\ncom.example.app second\n"
	s := newSandbox(t, fakeFixture{
		Entries: []fakeEntry{{
			Domain: "com.example.app", Key: "first", ReadOutput: "one\n", TypeOutput: "Type is string\n",
		}},
		Failures: []fakeFailure{{
			Operation: "read", Domain: "com.example.app", Key: "second", ExitCode: 43, Stderr: "second read failed",
		}},
	})
	s.writeConfig(t, original)

	result := s.run(t, "", "pull", "-y")

	requireExitCode(t, result, 1)
	if actual := s.readConfig(t); actual != original {
		t.Fatalf("partial pull changed config: want %q, got %q", original, actual)
	}
	s.requireInvocations(t,
		[]string{"read", "com.example.app", "first"},
		[]string{"read-type", "com.example.app", "first"},
		[]string{"read", "com.example.app", "second"},
	)
}

func TestPullConfigWriteFailureDoesNotReportSuccess(t *testing.T) {
	original := "com.example.app greeting\n"
	s := newSandbox(t, fakeFixture{Entries: []fakeEntry{{
		Domain: "com.example.app", Key: "greeting", ReadOutput: "hello\n", TypeOutput: "Type is string\n",
	}}})
	s.writeConfig(t, original)
	if err := os.Chmod(s.home, 0o500); err != nil {
		t.Fatalf("make config directory read-only: %v", err)
	}
	t.Cleanup(func() {
		_ = os.Chmod(s.home, 0o700)
	})

	result := s.run(t, "", "pull", "-y")

	requireExitCode(t, result, 1)
	requireNotContains(t, result.stdout+result.stderr, "Success:")
	if actual := s.readConfig(t); actual != original {
		t.Fatalf("write failure changed config: want %q, got %q", original, actual)
	}
	s.requireInvocations(t,
		[]string{"read", "com.example.app", "greeting"},
		[]string{"read-type", "com.example.app", "greeting"},
	)
}

func TestUnknownCommand(t *testing.T) {
	s := newSandbox(t, fakeFixture{})

	result := s.run(t, "", "not-a-command")

	requireExitCode(t, result, 1)
	combined := result.stdout + result.stderr + s.applicationLog(t)
	requireContainsFold(t, combined, "unknown command")
	s.requireInvocations(t)
}

func TestEntrypointFilesystemBehavior(t *testing.T) {
	t.Run("application log creation failure", func(t *testing.T) {
		s := newSandbox(t, fakeFixture{})
		if err := os.Chmod(s.work, 0o500); err != nil {
			t.Fatalf("make work directory read-only: %v", err)
		}

		result := s.run(t, "", "--version")

		requireExitCode(t, result, 0)
		requireContains(t, result.stderr, "Failed to open log file")
		requireContains(t, result.stdout, "Version: e2e-version")
		s.requireInvocations(t)
	})

	t.Run("unreadable config permissions are repaired", func(t *testing.T) {
		s := newSandbox(t, fakeFixture{})
		s.writeConfig(t, configLine("com.example.app", "key", "value", "string"))
		if err := os.Chmod(s.configPath, 0o000); err != nil {
			t.Fatalf("make config unreadable: %v", err)
		}
		t.Cleanup(func() {
			_ = os.Chmod(s.configPath, 0o600)
		})

		result := s.run(t, "", "config")

		requireExitCode(t, result, 0)
		requireContains(t, result.stdout, "com.example.app key value string")
		info, err := os.Stat(s.configPath)
		if err != nil {
			t.Fatalf("stat repaired config: %v", err)
		}
		if actual := info.Mode().Perm(); actual != 0o600 {
			t.Fatalf("expected repaired config permissions, got %o", actual)
		}
		s.requireInvocations(t)
	})
}

func newSandbox(t *testing.T, fixture fakeFixture) *sandbox {
	t.Helper()

	root := t.TempDir()
	home := filepath.Join(root, "home")
	work := filepath.Join(root, "work")
	fakeBin := filepath.Join(root, "fake-bin")
	for _, path := range []string{home, work, fakeBin} {
		if err := os.Mkdir(path, 0o700); err != nil {
			t.Fatalf("create sandbox directory %s: %v", path, err)
		}
	}

	testBinary, err := os.Executable()
	if err != nil {
		t.Fatalf("locate E2E test binary: %v", err)
	}
	if err := os.Symlink(testBinary, filepath.Join(fakeBin, "defaults")); err != nil {
		t.Fatalf("create fake defaults symlink: %v", err)
	}

	fixturePath := filepath.Join(root, "fixture.json")
	fixtureJSON, err := json.MarshalIndent(fixture, "", "  ")
	if err != nil {
		t.Fatalf("encode fake defaults fixture: %v", err)
	}
	if err := os.WriteFile(fixturePath, fixtureJSON, 0o600); err != nil {
		t.Fatalf("write fake defaults fixture: %v", err)
	}

	binaryPath := os.Getenv(binaryEnv)
	if binaryPath == "" {
		t.Fatalf("missing %s", binaryEnv)
	}
	if info, err := os.Stat(binaryPath); err != nil {
		t.Fatalf("stat mdefaults binary: %v", err)
	} else if info.Mode()&0o111 == 0 {
		t.Fatalf("mdefaults binary is not executable: %s", binaryPath)
	}

	return &sandbox{
		root:           root,
		home:           home,
		work:           work,
		fakeBin:        fakeBin,
		configPath:     filepath.Join(home, ".mdefaults"),
		fixturePath:    fixturePath,
		invocationPath: filepath.Join(root, "defaults-invocations.jsonl"),
		binaryPath:     binaryPath,
	}
}

func (s *sandbox) run(t *testing.T, stdin string, args ...string) commandResult {
	t.Helper()

	coverageDir := os.Getenv(coverageDirEnv)
	cmd := exec.Command(s.binaryPath, args...)
	cmd.Dir = s.work
	cmd.Env = []string{
		"GOCOVERDIR=" + coverageDir,
		"HOME=" + s.home,
		"MDEFAULTS_E2E_CONTAINER=1",
		fakeDefaultsEnv + "=1",
		fixtureEnv + "=" + s.fixturePath,
		invocationEnv + "=" + s.invocationPath,
		"NO_COLOR=1",
		"PATH=" + s.fakeBin,
		"TMPDIR=" + s.root,
	}
	cmd.Stdin = strings.NewReader(stdin)
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	exitCode := 0
	if err != nil {
		var exitError *exec.ExitError
		if !errors.As(err, &exitError) {
			t.Fatalf("run mdefaults: %v", err)
		}
		exitCode = exitError.ExitCode()
	}
	return commandResult{exitCode: exitCode, stdout: stdout.String(), stderr: stderr.String()}
}

func (s *sandbox) writeConfig(t *testing.T, content string) {
	t.Helper()
	if err := os.WriteFile(s.configPath, []byte(content), 0o600); err != nil {
		t.Fatalf("write sandbox config: %v", err)
	}
}

func (s *sandbox) readConfig(t *testing.T) string {
	t.Helper()
	content, err := os.ReadFile(filepath.Join(s.home, ".mdefaults"))
	if err != nil {
		t.Fatalf("read sandbox config: %v", err)
	}
	return string(content)
}

func (s *sandbox) applicationLog(t *testing.T) string {
	t.Helper()
	content, err := os.ReadFile(filepath.Join(s.work, "mdefaults.log"))
	if errors.Is(err, os.ErrNotExist) {
		return ""
	}
	if err != nil {
		t.Fatalf("read application log: %v", err)
	}
	return string(content)
}

func (s *sandbox) requireInvocations(t *testing.T, expected ...[]string) {
	t.Helper()
	actual := s.invocations(t)
	if !reflect.DeepEqual(actual, expected) {
		t.Fatalf("unexpected defaults invocations:\nwant %#v\n got %#v", expected, actual)
	}
}

func (s *sandbox) invocations(t *testing.T) [][]string {
	t.Helper()
	file, err := os.Open(s.invocationPath)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		t.Fatalf("open defaults invocation log: %v", err)
	}
	defer func() {
		if err := file.Close(); err != nil {
			t.Errorf("close defaults invocation log: %v", err)
		}
	}()

	var invocations [][]string
	decoder := json.NewDecoder(file)
	for {
		var invocation fakeInvocation
		if err := decoder.Decode(&invocation); errors.Is(err, io.EOF) {
			break
		} else if err != nil {
			t.Fatalf("decode defaults invocation: %v", err)
		}
		invocations = append(invocations, invocation.Args)
	}
	return invocations
}

func runFakeDefaults() int {
	args := os.Args[1:]
	if err := recordInvocation(os.Getenv(invocationEnv), args); err != nil {
		fmt.Fprintf(os.Stderr, "record fake defaults invocation: %v\n", err)
		return 70
	}

	fixtureJSON, err := os.ReadFile(os.Getenv(fixtureEnv))
	if err != nil {
		fmt.Fprintf(os.Stderr, "read fake defaults fixture: %v\n", err)
		return 70
	}
	var fixture fakeFixture
	if err := json.Unmarshal(fixtureJSON, &fixture); err != nil {
		fmt.Fprintf(os.Stderr, "decode fake defaults fixture: %v\n", err)
		return 70
	}
	if len(args) < 3 {
		fmt.Fprintf(os.Stderr, "invalid defaults arguments: %#v\n", args)
		return 64
	}

	operation, domain, key := args[0], args[1], args[2]
	for _, failure := range fixture.Failures {
		if failure.Operation == operation && failure.Domain == domain && failure.Key == key {
			if failure.Stderr != "" {
				fmt.Fprintln(os.Stderr, failure.Stderr)
			}
			if failure.ExitCode == 0 {
				return 1
			}
			return failure.ExitCode
		}
	}

	switch operation {
	case "read", "read-type":
		entry, ok := findFakeEntry(fixture.Entries, domain, key)
		if !ok {
			fmt.Fprintf(os.Stderr, "missing fake defaults entry for %s %s\n", domain, key)
			return 44
		}
		output := entry.ReadOutput
		if operation == "read-type" {
			output = entry.TypeOutput
		}
		if _, err := io.WriteString(os.Stdout, output); err != nil {
			fmt.Fprintf(os.Stderr, "write fake defaults output: %v\n", err)
			return 70
		}
		return 0
	case "write":
		if len(args) < 4 {
			fmt.Fprintf(os.Stderr, "invalid defaults write arguments: %#v\n", args)
			return 64
		}
		return 0
	default:
		fmt.Fprintf(os.Stderr, "unsupported fake defaults operation %q\n", operation)
		return 64
	}
}

func recordInvocation(path string, args []string) error {
	file, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o600)
	if err != nil {
		return err
	}
	if err := json.NewEncoder(file).Encode(fakeInvocation{Args: args}); err != nil {
		_ = file.Close()
		return err
	}
	return file.Close()
}

func findFakeEntry(entries []fakeEntry, domain, key string) (fakeEntry, bool) {
	for _, entry := range entries {
		if entry.Domain == domain && entry.Key == key {
			return entry, true
		}
	}
	return fakeEntry{}, false
}

func reportCoverage(coverageDir string) error {
	entries, err := os.ReadDir(coverageDir)
	if err != nil {
		return err
	}
	if len(entries) == 0 {
		return fmt.Errorf("no coverage data was produced")
	}

	cmd := exec.Command("go", "tool", "covdata", "percent", "-i="+coverageDir)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func configLine(domain, key, value, valueType string) string {
	return domain + " " + key + " " + strconv.Quote(value) + " " + valueType + "\n"
}

func requireExitCode(t *testing.T, result commandResult, expected int) {
	t.Helper()
	if result.exitCode != expected {
		t.Fatalf("unexpected exit code: want %d, got %d\nstdout:\n%s\nstderr:\n%s", expected, result.exitCode, result.stdout, result.stderr)
	}
}

func requireContains(t *testing.T, value, substring string) {
	t.Helper()
	if !strings.Contains(value, substring) {
		t.Fatalf("expected %q to contain %q", value, substring)
	}
}

func requireContainsFold(t *testing.T, value, substring string) {
	t.Helper()
	if !strings.Contains(strings.ToLower(value), strings.ToLower(substring)) {
		t.Fatalf("expected %q to contain %q ignoring case", value, substring)
	}
}

func requireNotContains(t *testing.T, value, substring string) {
	t.Helper()
	if strings.Contains(value, substring) {
		t.Fatalf("expected %q not to contain %q", value, substring)
	}
}

func requirePathNotExist(t *testing.T, path string) {
	t.Helper()
	if _, err := os.Stat(path); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("expected path not to exist: %s (stat error: %v)", path, err)
	}
}
