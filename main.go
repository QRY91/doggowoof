package main

import (
	"os"

	"github.com/QRY91/doggowoof/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
