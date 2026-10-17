package handler

import "github.com/bwmarrin/discordgo"

func (h *handler) handleApplicationComponent(
	session *discordgo.Session,
	interaction *discordgo.InteractionCreate,
) error {
	switch interaction.ApplicationCommandData().Name {
	case "anon":
		return h.interaction.AnonCommandInteraction(session, interaction)
	default:
		return session.InteractionRespond(interaction.Interaction,
			&discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "undefined command",
				},
			},
		)
	}
}
