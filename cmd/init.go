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
	Short: "🐾 Initialize DoggoWoof (setup database, config, start daemon)",
	Long: `Initialize DoggoWoof on your system:

- Creates local SQLite database
- Sets up configuration file
- Initializes alert sources
- Optionally starts daemon

Your guard dog will be ready to watch for alerts!`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("🐕 Initializing DoggoWoof...")

		// Get home directory
		home, err := os.UserHomeDir()
		if err != nil {
			fmt.Printf("❌ Error getting home directory: %v\n", err)
			return
		}

		// Create .doggowoof directory
		doggoDir := filepath.Join(home, ".doggowoof")
		if err := os.MkdirAll(doggoDir, 0755); err != nil {
			fmt.Printf("❌ Error creating .doggowoof directory: %v\n", err)
			return
		}

		// Create database file (placeholder for now)
		dbPath := filepath.Join(doggoDir, "alerts.db")
		if _, err := os.Stat(dbPath); os.IsNotExist(err) {
			file, err := os.Create(dbPath)
			if err != nil {
				fmt.Printf("❌ Error creating database: %v\n", err)
				return
			}
			file.Close()
			fmt.Printf("✅ Created database: %s\n", dbPath)
		} else {
			fmt.Printf("📂 Database already exists: %s\n", dbPath)
		}

		// Create config file
		configPath := filepath.Join(home, ".doggowoof.yaml")
		if _, err := os.Stat(configPath); os.IsNotExist(err) {
			defaultConfig := `# DoggoWoof Configuration
daemon:
  port: 8042
  webhook_endpoint: "/webhook"
  
notifications:
  desktop: true
  sound: true
  
ai:
  local_only: true
  training_threshold: 10
  
sources:
  discord: []
  github: []
  email: []
`
			if err := os.WriteFile(configPath, []byte(defaultConfig), 0644); err != nil {
				fmt.Printf("❌ Error creating config: %v\n", err)
				return
			}
			fmt.Printf("✅ Created config: %s\n", configPath)
		} else {
			fmt.Printf("📂 Config already exists: %s\n", configPath)
		}

		fmt.Println("\n🎉 DoggoWoof initialized successfully!")
		fmt.Println("\nNext steps:")
		fmt.Println("  doggo daemon start     # Start the background service")
		fmt.Println("  doggo watch --discord  # Add Discord monitoring")
		fmt.Println("  doggo status          # Check current state")

		// Desktop notification
		if err := beeep.Notify("DoggoWoof", "🐕 Your guard dog is ready!", ""); err != nil {
			// Ignore notification errors
		}
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
}
