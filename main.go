package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"discord-bot/device"
)

func main() {
	shellyIP := os.Getenv("SHELLY_IP")
	if shellyIP == "" {
		log.Fatal("Error: SHELLY_IP is not set")
	}

	fmt.Println("Shelly device IP:", shellyIP)

	shelly := device.NewShellySwitchPlus("shelly-1", "My Shelly Switch", shellyIP, 0)

	//
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	fmt.Println("Connecting to Shelly device...")
	if err := shelly.Connect(ctx); err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer shelly.Disconnect(ctx)

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
	fmt.Printf("Online:      %t\n", status.Online)
	fmt.Printf("Power:       %t\n", status.Power)
	fmt.Printf("Temperature: %.1f°C\n", status.Temperature)
	fmt.Printf("Voltage:     %s V\n", status.Metadata["voltage"])
	fmt.Printf("Current:     %s A\n", status.Metadata["current"])
	fmt.Printf("Active Power: %s W\n", status.Metadata["apower"])
}
