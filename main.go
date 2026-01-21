package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

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

	// Register the message handler BEFORE opening the connection
	dg.AddHandler(messageCreate)

	// Set intents to receive guild messages and message content
	// Note: You must enable "Message Content Intent" in the Discord Developer Portal
	// (Bot tab -> Privileged Gateway Intents -> Message Content Intent)
	dg.Identify.Intents = discordgo.IntentGuildMessages | discordgo.IntentMessageContent

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
	fmt.Println("Bot is now listening for messages. Press CTRL-C to exit.")

	// Wait for a termination signal
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc
}

// messageCreate is called every time a new message is created in a channel the bot has access to
func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	// Ignore messages from the bot itself
	if m.Author.ID == s.State.User.ID {
		return
	}

	// Respond to "ping" with "pong"
	if m.Content == "ping" {
		_, err := s.ChannelMessageSend(m.ChannelID, "pong 🏓")
		if err != nil {
			log.Printf("Error sending message: %v", err)
		}
	}
}
