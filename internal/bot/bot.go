package bot

import (
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

// Unregister slash commands on shutdown
func (b *Bot) unregisterCommands() error {
	registeredCommands, err := b.session.ApplicationCommands(b.session.State.User.ID, "")
	if err != nil {
		return err
	}

	for _, v := range registeredCommands {
		err := b.session.ApplicationCommandDelete(b.session.State.User.ID, "", v.ID)
		if err != nil {
			return err
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
