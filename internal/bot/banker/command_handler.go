package banker

import (
	"context"
	"fmt"
	"kaizen-hq/internal/bot/utils"
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
	FeatureID            = "banker"
)

func HandleCommand(ctx context.Context, s *discordgo.Session, i *discordgo.InteractionCreate) {
	requestID := ctx.Value("requestID").(string)

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
						Label: "Fulfill",
						Style: discordgo.SuccessButton,
						CustomID: utils.BuildCustomID(
							FeatureID,
							"fulfill",
							i.Member.User.ID, strconv.FormatInt(amount, 10),
						),
					},
					discordgo.Button{
						Label:    "Cancel",
						Style:    discordgo.DangerButton,
						CustomID: utils.BuildCustomID(FeatureID, "decline", i.Member.User.ID),
					},
				},
			},
		},
	})

	if err != nil {
		log.Printf("Error sending message to admin channel: %v", err)
	}
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
