package printer

import (
	"github.com/fatih/color"
)

// PrintError prints an error message in red color
func PrintError(message string) {
	color.Red("Error: %s", message)
}

// PrintSuccess prints a success message in green color
func PrintSuccess(message string) {
	color.Green("Success: %s", message)
}
