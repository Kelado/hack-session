package main

import (
	"fmt"
	"log"
	"os"
)

func main() {
	// Get the Shelly device IP from environment variable
	// Use dns-sd to discover it:
	//   dns-sd -B _shelly._tcp local.
	//   dns-sd -G v4 <device-name>.local
	shellyIP := os.Getenv("SHELLY_IP")
	if shellyIP == "" {
		log.Fatal("Error: SHELLY_IP is not set")
	}

	fmt.Println("Shelly device IP:", shellyIP)
	fmt.Println("Ready to connect to Shelly device...")
}
