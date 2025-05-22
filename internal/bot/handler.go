package bot

import (
	"context"
	"fmt"
	"kaizen-hq/internal/bot/banker"
	"log"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/google/uuid"
)

func HandleInteraction(s *discordgo.Session, i *discordgo.InteractionCreate) {
	requestID := uuid.New().String()

	ctx := context.WithValue(context.Background(), "requestID", requestID)

	switch i.Type {
	case discordgo.InteractionApplicationCommand:
		switch i.ApplicationCommandData().Name {
		case "ping":
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "Pong! I'm hana still under development",
				},
			})

		case "banker":
			banker.HandleCommand(ctx, s, i)

			// case "profile":
			// 	userProfile, err := b.profileRepository.GetProfileByDiscordID(context.Background(), i.Member.User.ID)
			// 	if err != nil {
			// 		fmt.Println(err)
			// 	}
			// 	var donatorStatus string
			// 	if userProfile.Donator == 0 {
			// 		donatorStatus = "False"
			// 	} else {
			// 		donatorStatus = "True"
			// 	}
			// 	// Create an embed as a container for profile information
			// 	embed := &discordgo.MessageEmbed{
			// 		Title: userProfile.Rank,
			// 		Description: "**" + userProfile.Name + "** (ID: " + strconv.Itoa(userProfile.TornID) + ")\n\n" +
			// 			"**Level:** " + strconv.Itoa(userProfile.Level) + "\n" +
			// 			"**Awards:** " + strconv.Itoa(userProfile.Awards) + "\n" +
			// 			"**Friends:** " + strconv.Itoa(userProfile.Friends) + "\n" +
			// 			"**Enemies:** " + strconv.Itoa(userProfile.Enemies) + "\n" +
			// 			"**Age:** " + strconv.Itoa(int(time.Since(userProfile.Signup).Hours())/24) + " days\n" +
			// 			"**Property:** " + userProfile.Property + "\n" +
			// 			"**Donator Status:** " + donatorStatus + "\n" +
			// 			"---\n" +
			// 			"**Activity:** 35minutes/day\n" +
			// 			"**Xanax Usage:** 2.3/day\n" +
			// 			"**Gym Energy:** 950/day\n",
			// 		Color: 0x800080,
			// 		Thumbnail: &discordgo.MessageEmbedThumbnail{
			// 			URL: userProfile.ProfileImage,
			// 		},
			// 		Footer: &discordgo.MessageEmbedFooter{
			// 			Text: "Profile as of " + time.Now().Format("January 2, 2006"),
			// 		},
			// 	}

			// 	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			// 		Type: discordgo.InteractionResponseChannelMessageWithSource,
			// 		Data: &discordgo.InteractionResponseData{
			// 			Embeds: []*discordgo.MessageEmbed{embed},
			// 		},
			// 	})
		}
	case discordgo.InteractionMessageComponent:
		customID := i.MessageComponentData().CustomID
		fmt.Println(customID)
		parts := strings.Split(customID, "_")
		if len(parts) == 0 {
			log.Println("Invalid custom ID format")
			return
		}

		feature := parts[0]
		switch feature {
		case "banker":
			banker.HandleComponent(ctx, s, i)
		}
	}
}
