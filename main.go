package main

import (
	"fmt"
	"log"
	"os"

	"github.com/bwmarrin/discordgo"
)

func main() {
	token := os.Getenv("DISCORD_TOKEN")
	if token == "" {
		log.Fatal("Error: DISCORD_TOKEN is not set")
	}

	// Channel ID where the bot will send the hello message
	// Enable Developer Mode in Discord (User Settings -> Advanced -> Developer Mode)
	// Then right-click a channel and "Copy ID"
	channelID := os.Getenv("DISCORD_CHANNEL_ID")
	if channelID == "" {
		log.Fatal("Error: DISCORD_CHANNEL_ID is not set")
	}

	dg, err := discordgo.New("Bot " + token)
	if err != nil {
		log.Fatalf("Error creating Discord session: %v", err)
	}

	err = dg.Open()
	if err != nil {
		log.Fatalf("Error opening connection: %v", err)
	}
	defer dg.Close()

	fmt.Println("Bot is running as:", dg.State.User.Username)

	// Send a "Hello, World!" message to the specified channel
	_, err = dg.ChannelMessageSend(channelID, "Hello, World! 👋")
	if err != nil {
		log.Fatalf("Error sending message: %v", err)
	}
	fmt.Println("Message sent successfully!")
}
