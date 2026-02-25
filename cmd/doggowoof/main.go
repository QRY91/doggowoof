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
	case "sniff":
		cli.RunSniff(os.Args[2:])
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
	fmt.Println(`doggowoof — OSINT coordination detection companion

Usage:
  doggo init                              Create config and alerts database
  doggo sniff [--source=X] [--detect=X] [--days=N]  Run detection heuristics
  doggo status                            Show dashboard (osmotic data + alerts)
  doggo alerts [--severity=X] [--ack=ID]  List or acknowledge alerts
  doggo bore                              Deep Bore agent monitoring status
  doggo bore --watch                      Watch bridge logs (long-running)
  doggo help                              Show this help
  doggo version                           Show version

Detection heuristics (--detect):
  temporal       Temporal clustering (accounts posting together)
  behavior       Behavior profiling (posting pattern analysis)
  coordination   Combined coordination scoring (default: all)

Guard dog for osmotic's OSINT data. Detects coordinated inauthentic behavior.
Also watches Doug so you don't have to.`)
}
