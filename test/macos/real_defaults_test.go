//go:build darwin && realdefaults

package macos_test

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

const (
	realDefaultsGuardEnv = "MDEFAULTS_RUN_REAL_DEFAULTS_SMOKE"
	realBinaryEnv        = "MDEFAULTS_REAL_BINARY"
	realDefaultsPath     = "/usr/bin/defaults"
)

type scalarCase struct {
	key              string
	value            string
	internalType     string
	seedTypeFlag     string
	acceptedMacTypes []string
}

func TestRealDefaultsScalarRoundTrip(t *testing.T) {
	if os.Getenv(realDefaultsGuardEnv) != "1" {
		t.Skipf("set %s=1 to run the destructive real-defaults smoke test", realDefaultsGuardEnv)
	}

	binaryPath := os.Getenv(realBinaryEnv)
	if binaryPath == "" {
		t.Fatalf("%s must point to the mdefaults binary under test", realBinaryEnv)
	}
	if !filepath.IsAbs(binaryPath) {
		t.Fatalf("%s must be an absolute path, got %q", realBinaryEnv, binaryPath)
	}
	if info, err := os.Stat(binaryPath); err != nil {
		t.Fatalf("stat mdefaults binary: %v", err)
	} else if info.Mode()&0o111 == 0 {
		t.Fatalf("mdefaults binary is not executable: %s", binaryPath)
	}
	if info, err := os.Stat(realDefaultsPath); err != nil {
		t.Fatalf("stat real defaults command: %v", err)
	} else if info.Mode()&0o111 == 0 {
		t.Fatalf("real defaults command is not executable: %s", realDefaultsPath)
	}

	root := t.TempDir()
	home := filepath.Join(root, "home")
	work := filepath.Join(root, "work")
	for _, path := range []string{home, work} {
		if err := os.Mkdir(path, 0o700); err != nil {
			t.Fatalf("create isolated directory %s: %v", path, err)
		}
	}
	environment := isolatedEnvironment(home, work)
	domain := uniqueDefaultsDomain(t)
	t.Cleanup(func() {
		cleanupDefaultsDomain(t, environment, work, domain)
	})

	testCases := []scalarCase{
		{key: "stringValue", value: "smoke value with spaces", internalType: "string", seedTypeFlag: "-string", acceptedMacTypes: []string{"string"}},
		{key: "integerValue", value: "42", internalType: "integer", seedTypeFlag: "-int", acceptedMacTypes: []string{"integer"}},
		{key: "booleanValue", value: "true", internalType: "boolean", seedTypeFlag: "-bool", acceptedMacTypes: []string{"boolean"}},
		{key: "floatValue", value: "3.5", internalType: "float", seedTypeFlag: "-float", acceptedMacTypes: []string{"float", "real"}},
		{key: "dateValue", value: "2026-07-11 01:23:45 +0000", internalType: "date", seedTypeFlag: "-date", acceptedMacTypes: []string{"date"}},
	}

	canonicalValues := make(map[string]string, len(testCases))
	for _, testCase := range testCases {
		mustRunDefaults(t, environment, work, "write", domain, testCase.key, testCase.seedTypeFlag, testCase.value)
		canonicalValues[testCase.key] = readDefault(t, environment, work, domain, testCase.key)
		requireDefaultType(t, environment, work, domain, testCase)
	}

	configPath := filepath.Join(home, ".mdefaults")
	var pullTargets strings.Builder
	for _, testCase := range testCases {
		fmt.Fprintf(&pullTargets, "%s %s\n", domain, testCase.key)
	}
	if err := os.WriteFile(configPath, []byte(pullTargets.String()), 0o600); err != nil {
		t.Fatalf("write isolated pull targets: %v", err)
	}

	mustRunMdefaults(t, binaryPath, environment, work, "pull", "-y")
	requirePulledConfig(t, configPath, domain, testCases, canonicalValues)

	mustRunDefaults(t, environment, work, "delete", domain)
	if _, err := runDefaults(environment, work, "read", domain); err == nil {
		t.Fatalf("defaults domain still exists after the pre-push delete: %s", domain)
	}

	mustRunMdefaults(t, binaryPath, environment, work, "push")
	for _, testCase := range testCases {
		requireDefaultType(t, environment, work, domain, testCase)
		actualValue := readDefault(t, environment, work, domain, testCase.key)
		if actualValue != canonicalValues[testCase.key] {
			t.Errorf("value changed after pull/push for %s: want %q, got %q", testCase.key, canonicalValues[testCase.key], actualValue)
		}
	}

	arrayTarget := domain + " arrayValue\n"
	mustRunDefaults(t, environment, work, "write", domain, "arrayValue", "-array", "one", "two")
	if err := os.WriteFile(configPath, []byte(arrayTarget), 0o600); err != nil {
		t.Fatalf("write unsupported array target: %v", err)
	}
	if output, err := runMdefaults(binaryPath, environment, work, "pull", "-y"); err == nil {
		t.Fatalf("unsupported array pull unexpectedly succeeded:\n%s", output)
	}
	if content, err := os.ReadFile(configPath); err != nil {
		t.Fatalf("read config after unsupported array pull: %v", err)
	} else if string(content) != arrayTarget {
		t.Fatalf("unsupported array pull changed config: want %q, got %q", arrayTarget, content)
	}
}

func uniqueDefaultsDomain(t *testing.T) string {
	t.Helper()
	randomBytes := make([]byte, 8)
	if _, err := rand.Read(randomBytes); err != nil {
		t.Fatalf("generate unique defaults domain: %v", err)
	}
	return fmt.Sprintf("com.fumiyakume.mdefaults.smoke.%d.%s", os.Getpid(), hex.EncodeToString(randomBytes))
}

func isolatedEnvironment(home, work string) []string {
	environment := make([]string, 0, len(os.Environ())+6)
	for _, entry := range os.Environ() {
		key, _, _ := strings.Cut(entry, "=")
		switch key {
		case "GOCOVERDIR", "HOME", "LANG", "LC_ALL", "NO_COLOR", "PATH", "TMPDIR":
			continue
		}
		if strings.HasPrefix(key, "MDEFAULTS_E2E_") {
			continue
		}
		environment = append(environment, entry)
	}
	return append(environment,
		"HOME="+home,
		"LANG=C",
		"LC_ALL=C",
		"NO_COLOR=1",
		"PATH=/usr/bin:/bin:/usr/sbin:/sbin",
		"TMPDIR="+work,
	)
}

func mustRunMdefaults(t *testing.T, binaryPath string, environment []string, work string, args ...string) {
	t.Helper()
	output, err := runMdefaults(binaryPath, environment, work, args...)
	if err != nil {
		t.Fatalf("mdefaults %q failed: %v\n%s", args, err, output)
	}
}

func runMdefaults(binaryPath string, environment []string, work string, args ...string) (string, error) {
	command := exec.Command(binaryPath, args...)
	command.Dir = work
	command.Env = environment
	output, err := command.CombinedOutput()
	return string(output), err
}

func mustRunDefaults(t *testing.T, environment []string, work string, args ...string) string {
	t.Helper()
	output, err := runDefaults(environment, work, args...)
	if err != nil {
		t.Fatalf("defaults %q failed: %v\n%s", args, err, output)
	}
	return output
}

func runDefaults(environment []string, work string, args ...string) (string, error) {
	command := exec.Command(realDefaultsPath, args...)
	command.Dir = work
	command.Env = environment
	output, err := command.CombinedOutput()
	return string(output), err
}

func readDefault(t *testing.T, environment []string, work, domain, key string) string {
	t.Helper()
	output := mustRunDefaults(t, environment, work, "read", domain, key)
	return strings.TrimSuffix(output, "\n")
}

func requireDefaultType(t *testing.T, environment []string, work, domain string, testCase scalarCase) {
	t.Helper()
	output := strings.TrimSpace(mustRunDefaults(t, environment, work, "read-type", domain, testCase.key))
	const prefix = "Type is "
	if !strings.HasPrefix(output, prefix) {
		t.Fatalf("unexpected defaults read-type output for %s: %q", testCase.key, output)
	}
	actualType := strings.ToLower(strings.TrimSpace(strings.TrimPrefix(output, prefix)))
	for _, acceptedType := range testCase.acceptedMacTypes {
		if actualType == acceptedType {
			return
		}
	}
	t.Errorf("unexpected defaults type for %s: want one of %v, got %q", testCase.key, testCase.acceptedMacTypes, actualType)
}

func requirePulledConfig(t *testing.T, path, domain string, testCases []scalarCase, expectedValues map[string]string) {
	t.Helper()
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read pulled config: %v", err)
	}
	lines := strings.Split(strings.TrimSuffix(string(content), "\n"), "\n")
	if len(lines) != len(testCases) {
		t.Fatalf("unexpected pulled config line count: want %d, got %d\n%s", len(testCases), len(lines), content)
	}

	for _, testCase := range testCases {
		prefix := domain + " " + testCase.key + " "
		suffix := " " + testCase.internalType
		line := findLineWithPrefix(lines, prefix)
		if line == "" {
			t.Errorf("missing pulled config line for %s", testCase.key)
			continue
		}
		if !strings.HasSuffix(line, suffix) {
			t.Errorf("pulled config has wrong type for %s: %q", testCase.key, line)
			continue
		}
		quotedValue := strings.TrimSuffix(strings.TrimPrefix(line, prefix), suffix)
		value, err := strconv.Unquote(quotedValue)
		if err != nil {
			t.Errorf("decode pulled config value for %s: %v", testCase.key, err)
			continue
		}
		if value != expectedValues[testCase.key] {
			t.Errorf("unexpected pulled value for %s: want %q, got %q", testCase.key, expectedValues[testCase.key], value)
		}
	}
}

func findLineWithPrefix(lines []string, prefix string) string {
	for _, line := range lines {
		if strings.HasPrefix(line, prefix) {
			return line
		}
	}
	return ""
}

func cleanupDefaultsDomain(t *testing.T, environment []string, work, domain string) {
	t.Helper()
	_, _ = runDefaults(environment, work, "delete", domain)
	output, err := runDefaults(environment, work, "read", domain)
	if err == nil {
		t.Errorf("failed to clean up defaults domain %s; remaining value:\n%s", domain, output)
		return
	}
	var exitError *exec.ExitError
	if !errors.As(err, &exitError) {
		t.Errorf("verify defaults cleanup for %s: %v", domain, err)
	}
}
