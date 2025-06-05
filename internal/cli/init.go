package cli

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	"github.com/gen2brain/beeep"
	_ "github.com/mattn/go-sqlite3"
	"github.com/spf13/cobra"
)

// initCmd represents the init command
var initCmd = &cobra.Command{
	Use:   "init",
	Short: "🐾 Initialize DOGGOWOOF (setup your guard dog!)",
	Long: `Initialize DOGGOWOOF on your system:

- Creates local SQLite database (YOUR DATA STAYS HOME!)
- Sets up configuration file (SMART DEFAULTS!)
- Initializes alert sources (READY TO WATCH!)
- Optionally starts daemon (GOOD BOY GOES TO WORK!)

Your guard dog will be ready to BARK at important alerts!`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("🚨🐕 INITIALIZING DOGGOWOOF... 🐕🚨")

		// Get home directory
		home, err := os.UserHomeDir()
		if err != nil {
			fmt.Printf("❌ WOOF! Error getting home directory: %v\n", err)
			return
		}

		// Create .doggowoof directory
		doggoDir := filepath.Join(home, ".doggowoof")
		if err := os.MkdirAll(doggoDir, 0755); err != nil {
			fmt.Printf("❌ WOOF! Error creating .doggowoof directory: %v\n", err)
			return
		}

		// Create and initialize SQLite database
		dbPath := filepath.Join(doggoDir, "alerts.db")
		if err := initializeDatabase(dbPath); err != nil {
			fmt.Printf("❌ WOOF! Error setting up database: %v\n", err)
			return
		}

		// Create config file
		configPath := filepath.Join(home, ".doggowoof.yaml")
		if _, err := os.Stat(configPath); os.IsNotExist(err) {
			defaultConfig := `# DOGGOWOOF Configuration - YOUR GUARD DOG SETTINGS!
daemon:
  port: 8042
  webhook_endpoint: "/webhook"
  
notifications:
  desktop: true    # BARK ON DESKTOP!
  sound: true      # LOUD WOOFS!
  
ai:
  local_only: true           # YOUR DATA STAYS HOME!
  training_threshold: 10     # HOW MANY ALERTS TO LEARN FROM
  
sources:
  discord: []      # DISCORD WEBHOOKS TO WATCH
  github: []       # GITHUB REPOS TO GUARD  
  email: []        # EMAIL SOURCES TO MONITOR
`
			if err := os.WriteFile(configPath, []byte(defaultConfig), 0644); err != nil {
				fmt.Printf("❌ WOOF! Error creating config: %v\n", err)
				return
			}
			fmt.Printf("✅ CREATED CONFIG: %s\n", configPath)
		} else {
			fmt.Printf("📂 CONFIG ALREADY EXISTS: %s\n", configPath)
		}

		// Copy daemon files to ~/.doggowoof/daemon/
		if err := copyDaemonFiles(doggoDir); err != nil {
			fmt.Printf("❌ WOOF! Error setting up daemon: %v\n", err)
			return
		}

		fmt.Println("\n🎉🚨 DOGGOWOOF INITIALIZED SUCCESSFULLY! 🚨🎉")
		fmt.Println("🐕 YOUR GUARD DOG IS READY TO PROTECT YOUR ALERTS! 🐕")
		fmt.Println("\nNEXT STEPS (UNLEASH THE POWER!):")
		fmt.Println("  doggo daemon start      # Start the guard dog service")
		fmt.Println("  doggo watch --github    # Add GitHub monitoring (WOOF AT CI!)")
		fmt.Println("  doggo daemon status     # Check if your good boy is awake")

		// Desktop notification
		if err := beeep.Notify("DOGGOWOOF", "🚨🐕 YOUR GUARD DOG IS READY! WOOF! 🐕🚨", ""); err != nil {
			// Ignore notification errors
		}
	},
}

// initializeDatabase creates and sets up the SQLite database
func initializeDatabase(dbPath string) error {
	// Check if database already exists
	if _, err := os.Stat(dbPath); err == nil {
		fmt.Printf("📂 DATABASE ALREADY EXISTS: %s\n", dbPath)
		return nil
	}

	// Create database
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return fmt.Errorf("failed to create database: %w", err)
	}
	defer db.Close()

	// Create alerts table (matches Python daemon schema)
	schema := `
	CREATE TABLE alerts (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		source TEXT NOT NULL,
		severity TEXT NOT NULL,
		title TEXT NOT NULL,
		message TEXT,
		raw_payload TEXT,
		timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
		acknowledged BOOLEAN DEFAULT FALSE
	);

	CREATE INDEX idx_alerts_source ON alerts(source);
	CREATE INDEX idx_alerts_severity ON alerts(severity);
	CREATE INDEX idx_alerts_timestamp ON alerts(timestamp);
	`

	if _, err := db.Exec(schema); err != nil {
		return fmt.Errorf("failed to create schema: %w", err)
	}

	fmt.Printf("✅ CREATED DATABASE: %s\n", dbPath)
	return nil
}

// copyDaemonFiles copies the Python daemon to the user's config directory
func copyDaemonFiles(doggoDir string) error {
	daemonDir := filepath.Join(doggoDir, "daemon")
	if err := os.MkdirAll(daemonDir, 0755); err != nil {
		return fmt.Errorf("failed to create daemon directory: %w", err)
	}

	// For now, we'll create a simple note about setting up the daemon manually
	// In a real deployment, we'd embed the Python files or download them
	setupNote := `# DOGGOWOOF Daemon Setup

The Python daemon files need to be copied here:
- main.py
- requirements.txt

For development, copy from the daemon/ directory in the source code.
For deployment, these would be embedded or downloaded automatically.

Manual setup:
1. Copy daemon/main.py to this directory
2. Copy daemon/requirements.txt to this directory  
3. Run: pip install -r requirements.txt
4. Run: doggo daemon start
`

	noteFile := filepath.Join(daemonDir, "SETUP.md")
	if err := os.WriteFile(noteFile, []byte(setupNote), 0644); err != nil {
		return fmt.Errorf("failed to create setup note: %w", err)
	}

	fmt.Printf("✅ CREATED DAEMON DIRECTORY: %s\n", daemonDir)
	fmt.Printf("📝 DAEMON SETUP INSTRUCTIONS: %s\n", noteFile)

	return nil
}

func init() {
	rootCmd.AddCommand(initCmd)
}
