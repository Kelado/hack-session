package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"discord-bot/device"
)

func printUsage() {
	fmt.Println("Usage: go run main.go [command]")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  status  - Show device status (default)")
	fmt.Println("  on      - Turn the switch ON")
	fmt.Println("  off     - Turn the switch OFF")
	fmt.Println("  toggle  - Toggle the switch state")
}

func main() {
	shellyIP := os.Getenv("SHELLY_IP")
	if shellyIP == "" {
		log.Fatal("Error: SHELLY_IP is not set")
	}

	// Get command from arguments (default: status)
	command := "status"
	if len(os.Args) > 1 {
		command = os.Args[1]
	}

	// Validate command
	validCommands := map[string]bool{"status": true, "on": true, "off": true, "toggle": true}
	if !validCommands[command] {
		printUsage()
		os.Exit(1)
	}

	shelly := device.NewShellySwitchPlus("shelly-1", "My Shelly Switch", shellyIP, 0)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	fmt.Println("Connecting to Shelly device...")
	if err := shelly.Connect(ctx); err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer shelly.Disconnect(ctx)

	// Execute the command
	switch command {
	case "on":
		fmt.Println("Turning switch ON...")
		if err := shelly.Execute(ctx, device.Command{Action: "on"}); err != nil {
			log.Fatalf("Failed to turn on: %v", err)
		}
		fmt.Println("✅ Switch is now ON")

	case "off":
		fmt.Println("Turning switch OFF...")
		if err := shelly.Execute(ctx, device.Command{Action: "off"}); err != nil {
			log.Fatalf("Failed to turn off: %v", err)
		}
		fmt.Println("✅ Switch is now OFF")

	case "toggle":
		fmt.Println("Toggling switch...")
		if err := shelly.Execute(ctx, device.Command{Action: "toggle"}); err != nil {
			log.Fatalf("Failed to toggle: %v", err)
		}
		fmt.Println("✅ Switch toggled")

	case "status":
		info := shelly.Info()
		extInfo := shelly.ExtendedInfo()
		fmt.Println("\n--- Device Info ---")
		fmt.Printf("ID:       %s\n", info.ID)
		fmt.Printf("Name:     %s\n", info.Name)
		fmt.Printf("Model:    %s\n", info.Model)
		fmt.Printf("Firmware: %s\n", info.Firmware)
		fmt.Printf("MAC:      %s\n", extInfo.MAC)
		fmt.Printf("Gen:      %d\n", extInfo.Gen)

		status, err := shelly.GetStatus(ctx)
		if err != nil {
			log.Fatalf("Failed to get status: %v", err)
		}

		fmt.Println("\n--- Switch Status ---")
		fmt.Printf("Online:       %t\n", status.Online)
		fmt.Printf("Power:        %t\n", status.Power)
		fmt.Printf("Temperature:  %.1f°C\n", status.Temperature)
		fmt.Printf("Voltage:      %s V\n", status.Metadata["voltage"])
		fmt.Printf("Current:      %s A\n", status.Metadata["current"])
		fmt.Printf("Active Power: %s W\n", status.Metadata["apower"])
	}
}
