package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"discord-bot/device"

	"github.com/bwmarrin/discordgo"
)

// Global Shelly device instance
var shelly *device.ShellySwitchPlus

// Define the slash commands
var commands = []*discordgo.ApplicationCommand{
	{
		Name:        "on",
		Description: "Turn the Shelly switch ON",
	},
	{
		Name:        "off",
		Description: "Turn the Shelly switch OFF",
	},
	{
		Name:        "status",
		Description: "Get the current status of the Shelly switch",
	},
}

// Map command names to their handlers
var commandHandlers = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){
	"on": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := shelly.Execute(ctx, device.Command{Action: "on"}); err != nil {
			respondWithError(s, i, "Failed to turn on: "+err.Error())
			return
		}

		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "✅ Switch is now **ON**",
			},
		})
	},
	"off": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := shelly.Execute(ctx, device.Command{Action: "off"}); err != nil {
			respondWithError(s, i, "Failed to turn off: "+err.Error())
			return
		}

		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "✅ Switch is now **OFF**",
			},
		})
	},
	"status": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		status, err := shelly.GetStatus(ctx)
		if err != nil {
			respondWithError(s, i, "Failed to get status: "+err.Error())
			return
		}

		info := shelly.Info()
		powerState := "🔴 OFF"
		if status.Power {
			powerState = "🟢 ON"
		}

		content := fmt.Sprintf("**%s** (%s)\n\n"+
			"**Power:** %s\n"+
			"**Temperature:** %.1f°C\n"+
			"**Voltage:** %s V\n"+
			"**Active Power:** %s W",
			info.Name, info.Model,
			powerState,
			status.Temperature,
			status.Metadata["voltage"],
			status.Metadata["apower"])

		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: content,
			},
		})
	},
}

func respondWithError(s *discordgo.Session, i *discordgo.InteractionCreate, msg string) {
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "❌ " + msg,
		},
	})
}

func main() {
	// Check environment variables
	token := os.Getenv("DISCORD_TOKEN")
	if token == "" {
		log.Fatal("Error: DISCORD_TOKEN is not set")
	}

	shellyIP := os.Getenv("SHELLY_IP")
	if shellyIP == "" {
		log.Fatal("Error: SHELLY_IP is not set")
	}

	// Initialize Shelly device
	shelly = device.NewShellySwitchPlus("shelly-1", "My Shelly Switch", shellyIP, 0)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	fmt.Println("Connecting to Shelly device...")
	if err := shelly.Connect(ctx); err != nil {
		log.Fatalf("Failed to connect to Shelly: %v", err)
	}
	fmt.Printf("Connected to Shelly: %s (%s)\n", shelly.Info().Name, shelly.Info().Model)

	// Initialize Discord bot
	dg, err := discordgo.New("Bot " + token)
	if err != nil {
		log.Fatalf("Error creating Discord session: %v", err)
	}

	// Register the interaction handler for slash commands
	dg.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		if handler, ok := commandHandlers[i.ApplicationCommandData().Name]; ok {
			handler(s, i)
		}
	})

	err = dg.Open()
	if err != nil {
		log.Fatalf("Error opening Discord connection: %v", err)
	}
	defer dg.Close()

	fmt.Println("Discord bot running as:", dg.State.User.Username)

	// Register slash commands globally
	registeredCommands := make([]*discordgo.ApplicationCommand, len(commands))
	for i, cmd := range commands {
		registered, err := dg.ApplicationCommandCreate(dg.State.User.ID, "", cmd)
		if err != nil {
			log.Fatalf("Cannot create command '%s': %v", cmd.Name, err)
		}
		registeredCommands[i] = registered
		fmt.Printf("Registered command: /%s\n", cmd.Name)
	}

	fmt.Println("\n🎉 Bot is ready! Use /on, /off, /status in Discord.")
	fmt.Println("Press CTRL-C to exit.")

	// Wait for a termination signal
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc

	// Cleanup
	fmt.Println("\nShutting down...")
	for _, cmd := range registeredCommands {
		dg.ApplicationCommandDelete(dg.State.User.ID, "", cmd.ID)
	}
	shelly.Disconnect(context.Background())
	fmt.Println("Goodbye!")
}
