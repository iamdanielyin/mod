package mod

import (
	"fmt"
	"github.com/iamdanielyin/mod"
)

func main() {
	// Test mod.yml auto-detection and loading
	app := mod.New()

	// Get the loaded mod configuration
	if modConfig := app.GetModConfig(); modConfig != nil {
		fmt.Printf("✅ Loaded mod.yml configuration:\n")
		fmt.Printf("  App Name: %s\n", modConfig.App.Name)
		fmt.Printf("  Display Name: %s\n", modConfig.App.DisplayName)
		fmt.Printf("  Description: %s\n", modConfig.App.Description)
		fmt.Printf("  Version: %s\n", modConfig.App.Version)
		fmt.Printf("  Port: %d\n", modConfig.Settings.Port)
		fmt.Printf("  Cache Strategy: %s\n", modConfig.Settings.CacheStrategy)

		// Show BigCache config if enabled
		if modConfig.Cache.BigCache.Enabled {
			fmt.Printf("  BigCache: Enabled (Shards: %d)\n", modConfig.Cache.BigCache.Shards)
		}

		// Show logging config
		if modConfig.Logging.Console.Enabled {
			fmt.Printf("  Console Logging: %s level\n", modConfig.Logging.Console.Level)
		}
	} else {
		fmt.Println("❌ No mod.yml configuration found")
		fmt.Println("💡 Try creating a mod.yml file or setting MOD_PATH environment variable")
	}

	fmt.Println("\n🚀 Starting server...")
	// Note: In a real application, you might want to use the port from config
	// app.Run(":8080")
}
