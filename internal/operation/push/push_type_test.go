package push

import (
	"bytes"
	"log"
	"os"
	"strings"
	"testing"

	"github.com/fumiya-kume/mdefaults/internal/config"
)

func TestPushWithTypes(t *testing.T) {
	var logOutput bytes.Buffer
	log.SetOutput(&logOutput)
	defer log.SetOutput(os.Stderr)

	value1 := "1"
	value2 := "2"
	configs := []config.Config{
		{Domain: "com.apple.dock", Key: "autohide", Value: &value1, Type: "boolean"},
		{Domain: "com.apple.trackpad", Key: "ClickThreshold", Value: &value2, Type: "integer"},
	}

	Push(configs)

	logStr := logOutput.String()
	if strings.Contains(logStr, "Failed to write") && !strings.Contains(logStr, "exit status 127") {
		t.Errorf("Unexpected error in log output: %s", logStr)
	}
}

func TestPushWithStringType(t *testing.T) {
	var logOutput bytes.Buffer
	log.SetOutput(&logOutput)
	defer log.SetOutput(os.Stderr)

	value := "test"
	configs := []config.Config{
		{Domain: "com.apple.dock", Key: "test", Value: &value, Type: "string"},
	}

	Push(configs)

	logStr := logOutput.String()
	if strings.Contains(logStr, "Failed to write") && !strings.Contains(logStr, "exit status 127") {
		t.Errorf("Unexpected error in log output: %s", logStr)
	}
}

func TestPushWithEmptyType(t *testing.T) {
	var logOutput bytes.Buffer
	log.SetOutput(&logOutput)
	defer log.SetOutput(os.Stderr)

	value := "test"
	configs := []config.Config{
		{Domain: "com.apple.dock", Key: "test", Value: &value, Type: ""},
	}

	Push(configs)

	logStr := logOutput.String()
	if strings.Contains(logStr, "Failed to write") && !strings.Contains(logStr, "exit status 127") {
		t.Errorf("Unexpected error in log output: %s", logStr)
	}
}
