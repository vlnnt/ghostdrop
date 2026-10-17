package handler

import "github.com/bwmarrin/discordgo"

var (
	commands = []*discordgo.ApplicationCommand{
		{
			Name:        "anon",
			Description: "Send an image anonymously.",
			Type:        discordgo.ChatApplicationCommand,
			Options: []*discordgo.ApplicationCommandOption{
				{
					Name:        "media",
					Description: "Image to send.",
					Type:        discordgo.ApplicationCommandOptionAttachment,
					Required:    true,
				},
				{
					Name:        "channel",
					Description: "Channel to send the image.",
					Type:        discordgo.ApplicationCommandOptionChannel,
					Required:    true,
					ChannelTypes: []discordgo.ChannelType{
						discordgo.ChannelTypeGuildText,
					},
				},
			},
		},
	}
)
