package bot

import (
	"github.com/bwmarrin/discordgo"
)

func GetCommands() []*discordgo.ApplicationCommand {
	return []*discordgo.ApplicationCommand{
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
	}
}
