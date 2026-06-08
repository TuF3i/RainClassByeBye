package main

import (
	"fmt"
	"os"

	"github.com/fatih/color"

	"RainClassByeBye/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, color.New(color.FgRed, color.Bold).Sprintf("error: %v", err))
		os.Exit(1)
	}
}
