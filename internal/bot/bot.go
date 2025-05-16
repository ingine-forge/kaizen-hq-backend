package bot

import (
	"context"
	"fmt"
	"kaizen-hq/internal/profile"
	"log"
	"strconv"
	"time"

	"github.com/bwmarrin/discordgo"
)

type Bot struct {
	session           *discordgo.Session
	profileRepository *profile.Repository
	commands          []*discordgo.ApplicationCommand
}

// List your commands here
var commands = []*discordgo.ApplicationCommand{
	{
		Name:        "ping",
		Description: "Replies with pong",
	},
	{
		Name:        "profile",
		Description: "Displays user data",
	},
	// Add more commands here if you want
}

func NewBot(token string, profileRepository *profile.Repository) (*Bot, error) {
	dg, err := discordgo.New("Bot " + token)
	if err != nil {
		return nil, err
	}

	bot := &Bot{
		session:           dg,
		commands:          commands,
		profileRepository: profileRepository,
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
	case "profile":
		userProfile, err := b.profileRepository.GetProfileByDiscordID(context.Background(), i.Member.User.ID)
		if err != nil {
			fmt.Println(err)
		}
		var donatorStatus string
		if userProfile.Donator == 0 {
			donatorStatus = "False"
		} else {
			donatorStatus = "True"
		}
		// Create an embed as a container for profile information
		embed := &discordgo.MessageEmbed{
			Title: userProfile.Rank,
			Description: "**" + userProfile.Name + "** (ID: " + strconv.Itoa(userProfile.TornID) + ")\n\n" +
				"**Level:** " + strconv.Itoa(userProfile.Level) + "\n" +
				"**Awards:** " + strconv.Itoa(userProfile.Awards) + "\n" +
				"**Friends:** " + strconv.Itoa(userProfile.Friends) + "\n" +
				"**Enemies:** " + strconv.Itoa(userProfile.Enemies) + "\n" +
				"**Age:** " + strconv.Itoa(int(time.Since(userProfile.Signup).Hours())/24) + " days\n" +
				"**Property:** " + userProfile.Property + "\n" +
				"**Donator Status:** " + donatorStatus + "\n" +
				"---\n" +
				"**Activity:** 35minutes/day\n" +
				"**Xanax Usage:** 2.3/day\n" +
				"**Gym Energy:** 950/day\n",
			Color: 0x800080,
			Thumbnail: &discordgo.MessageEmbedThumbnail{
				URL: userProfile.ProfileImage,
			},
			Footer: &discordgo.MessageEmbedFooter{
				Text: "Profile as of " + time.Now().Format("January 2, 2006"),
			},
		}

		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Embeds: []*discordgo.MessageEmbed{embed},
			},
		})
	}
}
