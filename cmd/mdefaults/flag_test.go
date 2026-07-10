package main

import "testing"

func TestParseArguments(t *testing.T) {
	testCases := []struct {
		name     string
		args     []string
		expected commandOptions
	}{
		{name: "empty"},
		{name: "command", args: []string{"pull"}, expected: commandOptions{command: "pull"}},
		{name: "long version", args: []string{"--version"}, expected: commandOptions{version: true}},
		{name: "short version", args: []string{"-v"}, expected: commandOptions{version: true}},
		{name: "flags before command", args: []string{"--verbose", "-y", "pull"}, expected: commandOptions{command: "pull", verbose: true, yes: true}},
		{name: "flags after command", args: []string{"pull", "--verbose", "-y"}, expected: commandOptions{command: "pull", verbose: true, yes: true}},
		{name: "boolean assignment", args: []string{"--verbose=false", "push"}, expected: commandOptions{command: "push"}},
		{name: "long help", args: []string{"--help"}, expected: commandOptions{help: true}},
		{name: "short help with command", args: []string{"config", "-h"}, expected: commandOptions{command: "config", help: true}},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual, err := parseArguments(tc.args)
			if err != nil {
				t.Fatalf("expected no error, got %v", err)
			}
			if actual != tc.expected {
				t.Fatalf("expected %+v, got %+v", tc.expected, actual)
			}
		})
	}
}

func TestParseArgumentsRejectsInvalidInput(t *testing.T) {
	testCases := []struct {
		name string
		args []string
	}{
		{name: "unknown flag", args: []string{"--unknown"}},
		{name: "single dash", args: []string{"-"}},
		{name: "multiple commands", args: []string{"pull", "push"}},
		{name: "flag terminator with argument", args: []string{"--", "pull"}},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := parseArguments(tc.args); err == nil {
				t.Fatal("expected an error")
			}
		})
	}
}
