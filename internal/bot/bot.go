package bot

import (
	"fmt"
	"log"

	"github.com/bwmarrin/discordgo"
)

type Bot struct {
	session  *discordgo.Session
	commands []*discordgo.ApplicationCommand
}

// List your commands here
var commands = []*discordgo.ApplicationCommand{
	{
		Name:        "ping",
		Description: "Replies with pong",
	},
	// Add more commands here if you want
}

func NewBot(token string) (*Bot, error) {
	dg, err := discordgo.New("Bot " + token)
	if err != nil {
		return nil, err
	}

	bot := &Bot{
		session:  dg,
		commands: commands,
	}

	dg.AddHandler(bot.handleInteraction)

	return bot, nil
}

func (b *Bot) Start() error {
	err := b.session.Open()
	if err != nil {
		return err
	}

	err = b.registerCommands()
	if err != nil {
		return err
	}

	return nil
}

func (b *Bot) Stop() error {
	err := b.unregisterCommands()
	if err != nil {
		log.Println("Error unregistering commands:", err)
	}
	return b.session.Close()
}

// Register slash commands on startup
func (b *Bot) registerCommands() error {
	for _, v := range b.commands {
		_, err := b.session.ApplicationCommandCreate(b.session.State.User.ID, "", v)
		if err != nil {
			return err
		}
	}
	return nil
}

// Unregister all slash commands on shutdown (both global and guild-specific)
func (b *Bot) unregisterCommands() error {
	// 1. Unregister global commands
	globalCommands, err := b.session.ApplicationCommands(b.session.State.User.ID, "")
	if err != nil {
		return fmt.Errorf("failed to fetch global commands: %w", err)
	}

	for _, cmd := range globalCommands {
		if err := b.session.ApplicationCommandDelete(b.session.State.User.ID, "", cmd.ID); err != nil {
			return fmt.Errorf("failed to delete global command '%s': %w", cmd.Name, err)
		}
	}

	// 2. Unregister guild-specific commands
	// Get all guilds the bot is a member of
	for _, guild := range b.session.State.Guilds {
		guildCommands, err := b.session.ApplicationCommands(b.session.State.User.ID, guild.ID)
		if err != nil {
			return fmt.Errorf("failed to fetch guild commands for %s: %w", guild.Name, err)
		}

		for _, cmd := range guildCommands {
			if err := b.session.ApplicationCommandDelete(b.session.State.User.ID, guild.ID, cmd.ID); err != nil {
				return fmt.Errorf("failed to delete guild command '%s' from guild '%s': %w",
					cmd.Name, guild.Name, err)
			}
		}
	}

	return nil
}

func (b *Bot) handleInteraction(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if i.Type != discordgo.InteractionApplicationCommand {
		return
	}

	switch i.ApplicationCommandData().Name {
	case "ping":
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "Pong! I'm hana still under development",
			},
		})
	}
}
