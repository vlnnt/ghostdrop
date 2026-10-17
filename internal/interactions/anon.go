package interactions

import (
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"
	"go.uber.org/zap"
)

func (i *Interactions) AnonCommandInteraction(
	session *discordgo.Session,
	interaction *discordgo.InteractionCreate,
) error {
	data := interaction.ApplicationCommandData()
	i.log.Info(
		"AnonCommandInteraction started",
		zap.String("guild_id", interaction.GuildID),
		zap.String("interaction_id", interaction.ID),
	)

	mediaOption := data.GetOption("media")
	if mediaOption == nil {
		err := fmt.Errorf("media option is required")
		i.log.Warn("Invalid anon command", zap.Error(err))

		if respondErr := session.InteractionRespond(
			interaction.Interaction,
			&discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: err.Error(),
					Flags:   discordgo.MessageFlagsEphemeral,
				},
			}); respondErr != nil {
			i.log.Error("Failed to respond to interaction", zap.Error(respondErr))
		}

		return err
	}

	channelOption := data.GetOption("channel")
	if channelOption == nil {
		err := fmt.Errorf("channel option is required")
		i.log.Warn("Invalid anon command", zap.Error(err))

		if respondErr := session.InteractionRespond(
			interaction.Interaction,
			&discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: err.Error(),
					Flags:   discordgo.MessageFlagsEphemeral,
				},
			}); respondErr != nil {
			i.log.Error("Failed to respond to interaction", zap.Error(respondErr))
		}

		return err
	}

	if data.Resolved == nil {
		err := fmt.Errorf("resolved data is empty")
		i.log.Warn("Invalid anon command", zap.Error(err))

		if respondErr := session.InteractionRespond(
			interaction.Interaction,
			&discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: err.Error(),
					Flags:   discordgo.MessageFlagsEphemeral,
				},
			}); respondErr != nil {
			i.log.Error("Failed to respond to interaction", zap.Error(respondErr))
		}

		return err
	}

	attachmentID, ok := mediaOption.Value.(string)
	if !ok {
		err := fmt.Errorf("invalid media attachment")
		i.log.Warn("Invalid anon command", zap.Error(err))

		if respondErr := session.InteractionRespond(
			interaction.Interaction,
			&discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: err.Error(),
					Flags:   discordgo.MessageFlagsEphemeral,
				},
			}); respondErr != nil {
			i.log.Error("Failed to respond to interaction", zap.Error(respondErr))
		}

		return err
	}

	media := data.Resolved.Attachments[attachmentID]
	if media == nil {
		err := fmt.Errorf("attachment not found")
		i.log.Warn(
			"Attachment not found",
			zap.String("attachment_id", attachmentID),
			zap.Error(err),
		)

		if respondErr := session.InteractionRespond(
			interaction.Interaction,
			&discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: err.Error(),
					Flags:   discordgo.MessageFlagsEphemeral,
				},
			}); respondErr != nil {
			i.log.Error("Failed to respond to interaction", zap.Error(respondErr))
		}

		return err
	}

	if media.ContentType == "" || !strings.HasPrefix(media.ContentType, "image/") {
		i.log.Warn(
			"Unsupported media type",
			zap.String("content_type", media.ContentType),
			zap.String("attachment_id", attachmentID),
		)

		if err := session.InteractionRespond(
			interaction.Interaction,
			&discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "⚠️ only images are allowed.",
					Flags:   discordgo.MessageFlagsEphemeral,
				},
			}); err != nil {
			i.log.Error("Failed to respond to interaction", zap.Error(err))
			return err
		}

		return nil
	}

	targetChannel := channelOption.ChannelValue(session)
	if targetChannel == nil {
		err := fmt.Errorf("target channel not found")
		i.log.Warn("Invalid target channel", zap.Error(err))

		if respondErr := session.InteractionRespond(
			interaction.Interaction,
			&discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: err.Error(),
					Flags:   discordgo.MessageFlagsEphemeral,
				},
			}); respondErr != nil {
			i.log.Error("Failed to respond to interaction", zap.Error(respondErr))
		}

		return err
	}

	i.log.Info(
		"Anon command validated",
		zap.String("channel_id", targetChannel.ID),
		zap.String("content_type", media.ContentType),
		zap.Int("file_size", media.Size),
	)

	if err := session.InteractionRespond(
		interaction.Interaction,
		&discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "⏳ sending anonymously.",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		},
	); err != nil {
		i.log.Error(
			"Failed to acknowledge anon command",
			zap.Error(err),
		)
		return err
	}

	embed := &discordgo.MessageEmbed{
		Title: "👻 Anonymous Drop",
		Image: &discordgo.MessageEmbedImage{
			URL: media.URL,
		},
	}

	message, err := session.ChannelMessageSendComplex(
		targetChannel.ID,
		&discordgo.MessageSend{
			Embeds: []*discordgo.MessageEmbed{embed},
			AllowedMentions: &discordgo.MessageAllowedMentions{
				Parse: []discordgo.AllowedMentionType{},
			},
		},
	)

	if err != nil {
		i.log.Error(
			"Failed to send anonymous image",
			zap.String("channel_id", targetChannel.ID),
			zap.Error(err),
		)

		contentSendError := "❌ failed to send image!"
		_, editErr := session.InteractionResponseEdit(
			interaction.Interaction,
			&discordgo.WebhookEdit{
				Content: &contentSendError,
			},
		)
		if editErr != nil {
			i.log.Error(
				"Failed to update interaction response",
				zap.Error(editErr),
			)
		}

		return err
	}

	i.log.Info(
		"Anonymous image sent successfully",
		zap.String("channel_id", targetChannel.ID),
		zap.String("message_id", message.ID),
	)

	content := "✅ sent anonymously!"
	_, err = session.InteractionResponseEdit(
		interaction.Interaction,
		&discordgo.WebhookEdit{
			Content: &content,
		},
	)
	if err != nil {
		i.log.Error(
			"Failed to update interaction response",
			zap.String("channel_id", targetChannel.ID),
			zap.Error(err),
		)
		return err
	}

	i.log.Info(
		"AnonCommandInteraction completed",
		zap.String("message_id", message.ID),
	)

	return nil
}
