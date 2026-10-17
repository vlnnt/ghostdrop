package handler

import "github.com/bwmarrin/discordgo"

var (
	commands = []*discordgo.ApplicationCommand{
		{
			Name:        "anon",
			Description: "Send an Image, GIF, video and audio to send anonymously.",
			Type:        discordgo.ChatApplicationCommand,
			Options: []*discordgo.ApplicationCommandOption{
				{
					Name:        "media",
					Description: "Image, GIF, video or audio to send anonymously.",
					Type:        discordgo.ApplicationCommandOptionAttachment,
					Required:    true,
				},
				{
					Name:        "channel",
					Description: "Channel to send the Image, GIF, video or audio to send anonymously.",
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
