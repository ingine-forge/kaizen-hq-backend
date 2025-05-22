package banker

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

// TODO: Work on this

func HandleComponent(ctx context.Context, s *discordgo.Session, i *discordgo.InteractionCreate) {
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
