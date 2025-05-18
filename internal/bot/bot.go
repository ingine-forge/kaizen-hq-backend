package bot

import (
	"context"
	"fmt"
	"kaizen-hq/internal/profile"
	"log"
	"math"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

// Constants
const (
	BankerAdminChannelID = "1373519539450544229" // Replace with your actual channel ID
	ProcessingTime       = 1 * time.Minute
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
	{
		Name:        "banker",
		Description: "Requests banker for the amount",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "amount",
				Description: "Amount to withdraw (5k, 1.5m, max, half, etc.)",
				Required:    true,
			},
		},
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
	dg.AddHandler(bot.handleInteractionMessages)

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

func (b *Bot) handleInteractionMessages(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if i.Type != discordgo.InteractionMessageComponent {
		return
	}

	handleComponentInteraction(s, i)
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
	case "banker":
		handleBankerCommand(s, i)
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

// handleBankerCommand processes the /banker command
func handleBankerCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
	// Extract the amount string
	options := i.ApplicationCommandData().Options
	optionMap := make(map[string]*discordgo.ApplicationCommandInteractionDataOption, len(options))
	for _, opt := range options {
		optionMap[opt.Name] = opt
	}

	amountStr := optionMap["amount"].StringValue()

	// Determine maxValue (replace with actual logic to get user's max amount)
	maxValue := int64(10000000) // Example value

	// Parse the amount
	amount, err := parseAmount(amountStr, maxValue)
	if err != nil {
		// Send ephemeral error message to user
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: fmt.Sprintf("Error: %s", err.Error()),
				Flags:   1 << 6, // Ephemeral flag - only the user can see it
			},
		})
		return
	}

	// Format amount with commas
	p := message.NewPrinter(language.English)
	formattedAmount := p.Sprintf("%d", amount)

	// Create a unique ID for this request
	requestID := fmt.Sprintf("banker_%d", time.Now().Unix())

	// Send ephemeral response to the requesting user
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: fmt.Sprintf("Got it! Your request for $%s is in the system now. Please give it about 15 minutes to process before making another one. I’m here if you need anything else!", formattedAmount),
			Flags:   1 << 6, // Ephemeral flag
		},
	})

	// Send request to admin channel with buttons
	_, err = s.ChannelMessageSendComplex(BankerAdminChannelID, &discordgo.MessageSend{
		Embeds: []*discordgo.MessageEmbed{
			{
				Title:       "New Banker Request",
				Description: "A user has requested a banker withdrawal",
				Color:       0x00BFFF, // Blue color for pending
				Fields: []*discordgo.MessageEmbedField{
					{
						Name:   "User",
						Value:  fmt.Sprintf("<@%s> (%s)", i.Member.User.ID, i.Member.User.Username),
						Inline: true,
					},
					{
						Name:   "Amount",
						Value:  formattedAmount,
						Inline: true,
					},
					{
						Name:   "Request Time",
						Value:  time.Now().Format(time.RFC1123),
						Inline: false,
					},
				},
				Footer: &discordgo.MessageEmbedFooter{
					Text: fmt.Sprintf("Request ID: %s", requestID),
				},
			},
		},
		Components: []discordgo.MessageComponent{
			discordgo.ActionsRow{
				Components: []discordgo.MessageComponent{
					discordgo.Button{
						Label:    "Fulfill",
						Style:    discordgo.SuccessButton,
						CustomID: fmt.Sprintf("fulfill_%s_%s_%d", requestID, i.Member.User.ID, amount),
					},
					discordgo.Button{
						Label:    "Cancel",
						Style:    discordgo.DangerButton,
						CustomID: fmt.Sprintf("cancel_%s_%s", requestID, i.Member.User.ID),
					},
				},
			},
		},
	})

	if err != nil {
		log.Printf("Error sending message to admin channel: %v", err)
	}
}

// handleComponentInteraction processes button clicks
func handleComponentInteraction(s *discordgo.Session, i *discordgo.InteractionCreate) {
	// Parse the custom ID
	customID := i.MessageComponentData().CustomID
	fmt.Println(customID)
	parts := strings.Split(customID, "_")

	if len(parts) < 2 {
		return
	}

	action := parts[0]
	requestID := parts[1]
	userID := parts[3]

	// Respond to the interaction immediately to prevent timeout
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseUpdateMessage,
		Data: &discordgo.InteractionResponseData{
			Content:    "",
			Components: []discordgo.MessageComponent{}, // Remove the buttons
			Embeds:     i.Message.Embeds,               // Keep the existing embed
		},
	})

	// Update the embed based on the action
	var updatedEmbed *discordgo.MessageEmbed
	if len(i.Message.Embeds) > 0 {
		updatedEmbed = i.Message.Embeds[0]

		if action == "fulfill" {
			// Extract amount from custom ID
			var amount int64
			if len(parts) >= 4 {
				amount, _ = strconv.ParseInt(parts[4], 10, 64)
			}

			updatedEmbed.Title = "Banker Request - In Progress"
			updatedEmbed.Color = 0xFFAA00 // Orange for in progress
			updatedEmbed.Fields = append(updatedEmbed.Fields, &discordgo.MessageEmbedField{
				Name:   "Status",
				Value:  "In Progress",
				Inline: false,
			})
			updatedEmbed.Fields = append(updatedEmbed.Fields, &discordgo.MessageEmbedField{
				Name:   "Handled By",
				Value:  i.Member.User.Username,
				Inline: false,
			})

			// Edit message with updated embed
			_, err := s.ChannelMessageEditComplex(&discordgo.MessageEdit{
				Channel: i.ChannelID,
				ID:      i.Message.ID,
				Embeds:  &[]*discordgo.MessageEmbed{updatedEmbed},
			})

			if err != nil {
				log.Printf("Error updating message: %v", err)
			}

			// Send DM to the requesting user
			channel, err := s.UserChannelCreate(userID)
			if err != nil {
				log.Printf("Error creating DM channel: %v", err)
				return
			}

			// Format amount with commas for DM
			p := message.NewPrinter(language.English)
			formattedAmount := p.Sprintf("$%d", amount)

			// Send "in progress" DM to user
			_, err = s.ChannelMessageSendEmbed(channel.ID, &discordgo.MessageEmbed{
				Title:       "Your Banker Request is Being Processed",
				Description: "Hey there! Your request is in the works and should be ready soon. Please hang tight while we process it.",
				Color:       0xFFAA00, // Orange for in progress
				Fields: []*discordgo.MessageEmbedField{
					{
						Name:   "Amount",
						Value:  formattedAmount,
						Inline: true,
					},
					{
						Name:   "Status",
						Value:  "In Progress",
						Inline: true,
					},
					{
						Name:   "Handler",
						Value:  fmt.Sprintf("<@%s>", userID),
						Inline: true,
					},
				},
				Footer: &discordgo.MessageEmbedFooter{
					Text: fmt.Sprintf("Request ID: %s - We'll notify you once it's done!", requestID),
				},
				Timestamp: time.Now().Format(time.RFC3339),
			})

			if err != nil {
				log.Printf("Error sending DM: %v", err)
			}

			// Start processing (simulate with wait time)
			go processBankerRequest(s, channel.ID, requestID, amount)

		} else if action == "cancel" {
			updatedEmbed.Title = "Banker Request - Cancelled"
			updatedEmbed.Color = 0xFF0000 // Red for cancelled
			updatedEmbed.Fields = append(updatedEmbed.Fields, &discordgo.MessageEmbedField{
				Name:   "Status",
				Value:  "Cancelled",
				Inline: false,
			})
			updatedEmbed.Fields = append(updatedEmbed.Fields, &discordgo.MessageEmbedField{
				Name:   "Cancelled By",
				Value:  i.Member.User.Username,
				Inline: false,
			})

			// Edit message with updated embed
			_, err := s.ChannelMessageEditComplex(&discordgo.MessageEdit{
				Channel: i.ChannelID,
				ID:      i.Message.ID,
				Embeds:  &[]*discordgo.MessageEmbed{updatedEmbed},
			})

			if err != nil {
				log.Printf("Error updating message: %v", err)
			}

			// Send DM to the requesting user
			channel, err := s.UserChannelCreate(userID)
			if err != nil {
				log.Printf("Error creating DM channel: %v", err)
				return
			}

			// Send "cancelled" DM to user
			_, err = s.ChannelMessageSendEmbed(channel.ID, &discordgo.MessageEmbed{
				Title:       "Banker Request - Cancelled",
				Description: "Your banker request has been cancelled",
				Color:       0xFF0000, // Red for cancelled
				Footer: &discordgo.MessageEmbedFooter{
					Text: fmt.Sprintf("Request ID: %s", requestID),
				},
				Timestamp: time.Now().Format(time.RFC3339),
			})

			if err != nil {
				log.Printf("Error sending DM: %v", err)
			}
		}
	}
}

// processBankerRequest simulates processing the request and updates the user
func processBankerRequest(s *discordgo.Session, channelID string, requestID string, amount int64) {
	// Wait for the processing time
	time.Sleep(ProcessingTime)

	// Simulate a success/failure (90% success rate for example)
	success := (time.Now().UnixNano() % 10) < 9

	// Format amount with commas
	p := message.NewPrinter(language.English)
	formattedAmount := p.Sprintf("%d", amount)

	// Send final status message
	var finalEmbed *discordgo.MessageEmbed
	if success {
		finalEmbed = &discordgo.MessageEmbed{
			Title:       "Banker Request - Completed",
			Description: "Your banker request has been completed successfully",
			Color:       0x00FF00, // Green for success
			Fields: []*discordgo.MessageEmbedField{
				{
					Name:   "Amount",
					Value:  formattedAmount,
					Inline: true,
				},
				{
					Name:   "Status",
					Value:  "Completed",
					Inline: true,
				},
				{
					Name:   "Completion Time",
					Value:  time.Now().Format(time.RFC1123),
					Inline: false,
				},
			},
			Footer: &discordgo.MessageEmbedFooter{
				Text: fmt.Sprintf("Request ID: %s", requestID),
			},
		}
	} else {
		finalEmbed = &discordgo.MessageEmbed{
			Title:       "Banker Request - Failed",
			Description: "Unfortunately, your banker request could not be completed",
			Color:       0xFF0000, // Red for failure
			Fields: []*discordgo.MessageEmbedField{
				{
					Name:   "Amount",
					Value:  formattedAmount,
					Inline: true,
				},
				{
					Name:   "Status",
					Value:  "Failed",
					Inline: true,
				},
				{
					Name:   "Reason",
					Value:  "System error. Please try again later or contact an administrator.",
					Inline: false,
				},
			},
			Footer: &discordgo.MessageEmbedFooter{
				Text: fmt.Sprintf("Request ID: %s", requestID),
			},
		}
	}

	// Send the final status message to the user
	_, err := s.ChannelMessageSendEmbed(channelID, finalEmbed)
	if err != nil {
		log.Printf("Error sending final status DM: %v", err)
	}

	// Update the admin message as well (if needed)
	// This would require storing the admin channel message ID somewhere
}

// parseAmount converts strings like "5k", "1.5m", "half", etc. into integer values
func parseAmount(input string, maxValue int64) (int64, error) {
	input = strings.ToLower(strings.TrimSpace(input))

	// Handle special text cases
	switch input {
	case "max":
		return maxValue, nil
	case "half", "1/2", "50%":
		return int64(math.Floor(float64(maxValue) * 0.5)), nil
	case "quarter", "1/4", "25%":
		return int64(math.Floor(float64(maxValue) * 0.25)), nil
	case "1/3", "33%":
		return int64(math.Floor(float64(maxValue) * (1.0 / 3.0))), nil
	}

	// Handle percentage cases
	if strings.HasSuffix(input, "%") {
		percentStr := strings.TrimSuffix(input, "%")
		percent, err := strconv.ParseFloat(percentStr, 64)
		if err != nil {
			return 0, fmt.Errorf("invalid percentage format: %s", input)
		}
		return int64(math.Floor(float64(maxValue) * (percent / 100.0))), nil
	}

	// Handle fraction cases with regex
	fractionRegex := regexp.MustCompile(`^(\d+)/(\d+)$`)
	matches := fractionRegex.FindStringSubmatch(input)
	if len(matches) == 3 {
		numerator, err1 := strconv.ParseFloat(matches[1], 64)
		denominator, err2 := strconv.ParseFloat(matches[2], 64)
		if err1 != nil || err2 != nil || denominator == 0 {
			return 0, fmt.Errorf("invalid fraction format: %s", input)
		}
		return int64(math.Floor(float64(maxValue) * (numerator / denominator))), nil
	}

	// Handle numeric values with suffixes (k, m, b)
	numRegex := regexp.MustCompile(`^([-+]?\d*\.?\d+)([kmb])?$`)
	matches = numRegex.FindStringSubmatch(input)
	if len(matches) >= 2 {
		base, err := strconv.ParseFloat(matches[1], 64)
		if err != nil {
			return 0, fmt.Errorf("invalid number format: %s", input)
		}

		// Apply multiplier if suffix exists
		if len(matches) == 3 && matches[2] != "" {
			switch matches[2] {
			case "k":
				base *= 1000
			case "m":
				base *= 1000000
			case "b":
				base *= 1000000000
			}
		}

		// Convert to integer (floor)
		return int64(math.Floor(base)), nil
	}

	// If we get here, the format wasn't recognized
	return 0, fmt.Errorf("unrecognized amount format: %s", input)
}
