package main

import (
	"context"
	"fmt"
)

func main() {
	fs := &osFileSystem{}
	createConfigFileIfMissing(fs)
	content, err := readConfigFile(fs)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(string(content))

	defaults := &DefaultsCommandImpl{
		domain: "com.apple.dock",
		key:    "autohide",
	}

	result, err := defaults.Read(context.Background())
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(result)
}
