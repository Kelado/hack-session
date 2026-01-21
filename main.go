package main

import (
	"fmt"
	"log"

	"github.com/bwmarrin/discordgo"
)

func main() {
	// TODO: Replace with your actual Bot Token
	token := "REDACTED"

	// TODO: Replace with the Channel ID where you want to send the message
	channelID := "1336736337608315000"

	// 1. Create a session
	dg, err := discordgo.New("Bot " + token)
	if err != nil {
		log.Fatalf("Error creating Discord session: %v", err)
	}

	// 2. Open the connection
	err = dg.Open()
	if err != nil {
		log.Fatalf("Error opening connection: %v", err)
	}
	defer dg.Close()

	fmt.Println("Bot is connected!")

	// 3. Send a message
	_, err = dg.ChannelMessageSend(channelID, "Welcome to our hack session!")
	if err != nil {
		fmt.Printf("Error sending message: %v\n", err)
	} else {
		fmt.Println("Message sent successfully!")
	}
}
