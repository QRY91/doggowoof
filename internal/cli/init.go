package cli

import (
	"fmt"
	"os"

	"github.com/QRY91/doggowoof/internal/config"
	"github.com/QRY91/doggowoof/pkg/alerts"
	"github.com/gen2brain/beeep"
)

// RunInit creates the config file and initializes the alerts database.
func RunInit() {
	// Create config
	if err := config.Init(); err != nil {
		fatalf("Error: %v", err)
	}

	// Open alerts database to trigger creation + migrations
	cfg := config.DefaultConfig()
	dbPath := config.ExpandPath(cfg.Alerts.DatabasePath)
	store, err := alerts.New(dbPath)
	if err != nil {
		fatalf("Error creating alerts database: %v", err)
	}
	store.Close()

	fmt.Printf("Alerts database created at %s\n", dbPath)
	fmt.Println()

	// Check if osmotic database exists
	osmoticPath := config.ExpandPath(cfg.Osmotic.DatabasePath)
	if _, err := os.Stat(osmoticPath); os.IsNotExist(err) {
		fmt.Println("Note: osmotic database not found at", osmoticPath)
		fmt.Println("  Run 'osmotic init' and collect some data first.")
		fmt.Println("  Then use 'doggo sniff' to detect patterns.")
	} else {
		fmt.Println("Osmotic database found at", osmoticPath)
		fmt.Println()
		fmt.Println("Next steps:")
		fmt.Println("  doggo sniff                    Run all detection heuristics")
		fmt.Println("  doggo sniff --detect=temporal   Run temporal clustering only")
		fmt.Println("  doggo status                   Show dashboard")
	}

	// Desktop notification
	_ = beeep.Notify("DoggoWoof", "Guard dog initialized! Ready to sniff.", "")
}
