Steps
Setup Bot Application (Portal)

Create a New Application in the [Discord Developer Portal](https://discord.com/developers/applications).
Navigate to Bot -> Reset Token to get your Token (save this!).
Navigate to OAuth2 -> URL Generator.
Select Scope: bot.
Select Permissions: Send Messages.
Copy the URL, paste it in your browser, and invite the bot to your server.
Initialize Go Project

Create a folder (e.g., hack-session).
Run go mod init discord-bot.
Install the library: go get github.com/bwmarrin/discordgo.
Showcase 1: The "Broadcaster" (Send Only)

Create main.go that connects and sends a single "Hello" message to a hardcoded CHANNEL_ID (enable User Settings -> Advanced -> Developer Mode to copy ID).
Demonstrates: discordgo.New(), dg.Open(), dg.ChannelMessageSend().
Upgrade Permissions (Reading Actions)

Portal: Go to the Bot tab in Developer Portal.
Toggle On: Message Content Intent (Privileged Gateway Intents). Save changes.
Explain that modern Discord bots need special permission to read message text.
Showcase 2: The "Responder" (Interactive)

Update main.go to keep the connection open (make(chan os.Signal)).
Add dg.Identify.Intents = discordgo.IntentGuildMessages | discordgo.IntentMessageContent.
Register a handler: dg.AddHandler(func(s, m) { ... }).
Logic: If m.Content == "ping", call s.ChannelMessageSend(m.ChannelID, "pong").
