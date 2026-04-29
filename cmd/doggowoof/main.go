package main

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/QRY91/doggowoof/internal/cli"
)

func main() {
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: slog.LevelWarn,
	})))

	if len(os.Args) < 2 {
		printHelp()
		return
	}

	switch os.Args[1] {
	case "init":
		cli.RunInit()
	case "status":
		cli.RunStatus(os.Args[2:])
	case "alerts":
		cli.RunAlerts(os.Args[2:])
	case "bore":
		cli.RunBore(os.Args[2:])
	case "pet":
		cli.RunPet()
	case "help", "-h", "--help":
		printHelp()
	case "version", "--version":
		fmt.Println("doggowoof v0.2.0")
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n\n", os.Args[1])
		printHelp()
		os.Exit(1)
	}
}

func printHelp() {
	fmt.Println(`doggowoof — local-first AI agent observability

Usage:
  doggo init                              Create config and alerts database
  doggo status                            Show dashboard (active session + alerts)
  doggo bore                              Deep Bore agent monitoring status
  doggo bore --watch                      Watch bridge logs (long-running)
  doggo alerts [--severity=X] [--ack=ID]  List or acknowledge alerts
  doggo help                              Show this help
  doggo version                           Show version

Watches autonomous AI agents for sandbox violations, self-modification,
binary integrity changes, and config drift. Local-only, SQLite-backed,
desktop notifications when something matters. Good dog watches Doug.`)
}
