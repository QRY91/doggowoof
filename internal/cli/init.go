package cli

import (
	"fmt"

	"github.com/QRY91/doggowoof/internal/config"
	"github.com/QRY91/doggowoof/pkg/alerts"
	"github.com/gen2brain/beeep"
)

// RunInit creates the config file and initializes the alerts database.
func RunInit() {
	if err := config.Init(); err != nil {
		fatalf("Error: %v", err)
	}

	cfg := config.DefaultConfig()
	dbPath := config.ExpandPath(cfg.Alerts.DatabasePath)
	store, err := alerts.New(dbPath)
	if err != nil {
		fatalf("Error creating alerts database: %v", err)
	}
	store.Close()

	fmt.Printf("Alerts database created at %s\n", dbPath)
	fmt.Println()
	fmt.Println("Next steps:")
	fmt.Println("  Edit ~/.config/doggowoof/config.toml to enable Deep Bore monitoring.")
	fmt.Println("  Set [deepbore] enabled = true and adjust paths to your agent setup.")
	fmt.Println("  Then: doggo bore --watch")

	_ = beeep.Notify("DoggoWoof", "Guard dog initialized! Ready to watch.", "")
}
