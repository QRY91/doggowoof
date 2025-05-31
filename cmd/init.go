package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/gen2brain/beeep"
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

		// Create database file (placeholder for now)
		dbPath := filepath.Join(doggoDir, "alerts.db")
		if _, err := os.Stat(dbPath); os.IsNotExist(err) {
			file, err := os.Create(dbPath)
			if err != nil {
				fmt.Printf("❌ WOOF! Error creating database: %v\n", err)
				return
			}
			file.Close()
			fmt.Printf("✅ CREATED DATABASE: %s\n", dbPath)
		} else {
			fmt.Printf("📂 DATABASE ALREADY EXISTS: %s\n", dbPath)
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

		fmt.Println("\n🎉🚨 DOGGOWOOF INITIALIZED SUCCESSFULLY! 🚨🎉")
		fmt.Println("🐕 YOUR GUARD DOG IS READY TO PROTECT YOUR ALERTS! 🐕")
		fmt.Println("\nNEXT STEPS (UNLEASH THE POWER!):")
		fmt.Println("  doggo daemon start     # Start the guard dog service")
		fmt.Println("  doggo watch --discord  # Add Discord monitoring (WOOF AT MESSAGES!)")
		fmt.Println("  doggo status          # Check if your good boy is awake")

		// Desktop notification
		if err := beeep.Notify("DOGGOWOOF", "🚨🐕 YOUR GUARD DOG IS READY! WOOF! 🐕🚨", ""); err != nil {
			// Ignore notification errors
		}
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
}
