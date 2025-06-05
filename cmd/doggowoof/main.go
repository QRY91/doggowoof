package main

import (
	"os"

	"github.com/QRY91/doggowoof/internal/cli"
)

func main() {
	if err := cli.Execute(); err != nil {
		os.Exit(1)
	}
}
