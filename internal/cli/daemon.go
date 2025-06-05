package cli

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
)

// daemonCmd represents the daemon command
var daemonCmd = &cobra.Command{
	Use:   "daemon",
	Short: "🐕 Manage the DOGGOWOOF daemon (guard dog service)",
	Long: `Control the DOGGOWOOF webhook daemon:

The daemon is the heart of DOGGOWOOF - it listens for webhooks,
analyzes alerts, and sends notifications when something matters!

Commands:
  start    Start the guard dog daemon (WOOF!)
  stop     Stop the daemon (good boy, rest)
  restart  Restart the daemon (refresh the good boy)
  status   Check if daemon is running (is doggo awake?)`,
}

var daemonStartCmd = &cobra.Command{
	Use:   "start",
	Short: "🚨 Start the DOGGOWOOF daemon",
	Long:  `Start the webhook receiver daemon. Your guard dog goes to work!`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := startDaemon(); err != nil {
			fmt.Printf("❌ WOOF! Failed to start daemon: %v\n", err)
			return
		}
		fmt.Println("🚨🐕 DAEMON STARTED! Guard dog is now watching! 🐕🚨")
	},
}

var daemonStopCmd = &cobra.Command{
	Use:   "stop",
	Short: "🛑 Stop the DOGGOWOOF daemon",
	Long:  `Stop the webhook receiver daemon. Guard dog takes a break.`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := stopDaemon(); err != nil {
			fmt.Printf("❌ WOOF! Failed to stop daemon: %v\n", err)
			return
		}
		fmt.Println("🛑 DAEMON STOPPED. Good boy is resting. 🐕")
	},
}

var daemonStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "📊 Check daemon status",
	Long:  `Check if the DOGGOWOOF daemon is running and healthy.`,
	Run: func(cmd *cobra.Command, args []string) {
		status, err := getDaemonStatus()
		if err != nil {
			fmt.Printf("❌ WOOF! Failed to check status: %v\n", err)
			return
		}
		fmt.Println(status)
	},
}

func startDaemon() error {
	// Check if already running
	if isRunning, _ := isDaemonRunning(); isRunning {
		return fmt.Errorf("daemon is already running")
	}

	// Get daemon path
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}

	daemonPath := filepath.Join(home, ".doggowoof", "daemon")

	// Check if Python daemon exists
	pythonDaemon := filepath.Join(daemonPath, "main.py")
	if _, err := os.Stat(pythonDaemon); os.IsNotExist(err) {
		return fmt.Errorf("daemon not found at %s. Run 'doggo init' first", pythonDaemon)
	}

	// Start daemon in background
	cmd := exec.Command("python3", pythonDaemon)
	cmd.Dir = daemonPath

	// Redirect output to log files
	logDir := filepath.Join(home, ".doggowoof", "logs")
	os.MkdirAll(logDir, 0755)

	logFile, err := os.Create(filepath.Join(logDir, "daemon.log"))
	if err != nil {
		return fmt.Errorf("failed to create log file: %w", err)
	}

	cmd.Stdout = logFile
	cmd.Stderr = logFile

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start daemon: %w", err)
	}

	// Save PID for later management
	pidFile := filepath.Join(home, ".doggowoof", "daemon.pid")
	pidContent := strconv.Itoa(cmd.Process.Pid)
	if err := os.WriteFile(pidFile, []byte(pidContent), 0644); err != nil {
		return fmt.Errorf("failed to save PID: %w", err)
	}

	return nil
}

func stopDaemon() error {
	pid, err := getDaemonPID()
	if err != nil {
		return fmt.Errorf("daemon not running or PID file missing")
	}

	// Kill the process
	process, err := os.FindProcess(pid)
	if err != nil {
		return fmt.Errorf("failed to find process: %w", err)
	}

	if err := process.Kill(); err != nil {
		return fmt.Errorf("failed to kill process: %w", err)
	}

	// Remove PID file
	home, _ := os.UserHomeDir()
	pidFile := filepath.Join(home, ".doggowoof", "daemon.pid")
	os.Remove(pidFile)

	return nil
}

func getDaemonStatus() (string, error) {
	isRunning, pid := isDaemonRunning()

	if !isRunning {
		return "🔴 DAEMON: NOT RUNNING (doggo is sleeping)", nil
	}

	// Check if daemon is responsive
	// TODO: Add health check to daemon endpoint
	return fmt.Sprintf("🟢 DAEMON: RUNNING (PID: %d) - Guard dog is alert! 🐕", pid), nil
}

func isDaemonRunning() (bool, int) {
	pid, err := getDaemonPID()
	if err != nil {
		return false, 0
	}

	// Check if process exists
	process, err := os.FindProcess(pid)
	if err != nil {
		return false, 0
	}

	// On Unix, Signal with signal 0 tests if process exists
	err = process.Signal(os.Signal(nil))
	return err == nil, pid
}

func getDaemonPID() (int, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return 0, err
	}

	pidFile := filepath.Join(home, ".doggowoof", "daemon.pid")
	pidBytes, err := os.ReadFile(pidFile)
	if err != nil {
		return 0, err
	}

	pid, err := strconv.Atoi(strings.TrimSpace(string(pidBytes)))
	if err != nil {
		return 0, err
	}

	return pid, nil
}

func init() {
	rootCmd.AddCommand(daemonCmd)
	daemonCmd.AddCommand(daemonStartCmd)
	daemonCmd.AddCommand(daemonStopCmd)
	daemonCmd.AddCommand(daemonStatusCmd)
}
