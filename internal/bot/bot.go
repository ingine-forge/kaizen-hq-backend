package bot

import (
	"fmt"
	"log"
	"time"

	"github.com/bwmarrin/discordgo"
)

type Bot struct {
	session  *discordgo.Session
	commands []*discordgo.ApplicationCommand
}

func NewBot(token string) (*Bot, error) {
	dg, err := discordgo.New("Bot " + token)
	if err != nil {
		return nil, err
	}

	bot := &Bot{
		session:  dg,
		commands: GetCommands(),
	}

	dg.AddHandler(HandleInteraction)

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

// Register slash commands only at guild level for faster registration
func (b *Bot) registerCommands() error {
	// If no guilds are available yet, wait a moment for the bot to connect
	if len(b.session.State.Guilds) == 0 {
		time.Sleep(2 * time.Second)
	}

	// Get all guilds the bot is in
	guilds := b.session.State.Guilds
	if len(guilds) == 0 {
		return fmt.Errorf("bot is not in any guilds, cannot register commands")
	}

	// Register commands for each guild individually (not globally)
	for _, guild := range guilds {
		for _, cmd := range b.commands {
			_, err := b.session.ApplicationCommandCreate(b.session.State.User.ID, guild.ID, cmd)
			if err != nil {
				return fmt.Errorf("failed to register command '%s' in guild '%s': %w",
					cmd.Name, guild.Name, err)
			}
		}
		log.Printf("Registered %d commands in guild '%s'", len(b.commands), guild.Name)
	}

	log.Printf("Successfully registered commands in %d guilds", len(guilds))
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
