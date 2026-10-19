package interactions

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"go.uber.org/zap"
)

const maxMediaSize int64 = 20 << 20

var supportedMedia = map[string]string{
	"image/png":  ".png",
	"image/jpeg": ".jpg",
	"image/gif":  ".gif",
	"image/webp": ".webp",

	"video/mp4":       ".mp4",
	"video/webm":      ".webm",
	"video/quicktime": ".mov",

	"audio/mpeg":    ".mp3",
	"audio/mpeg3":   ".mp3",
	"audio/mp3":     ".mp3",
	"audio/x-mpeg":  ".mp3",
	"audio/x-mpeg3": ".mp3",
	"audio/x-mp3":   ".mp3",
	"audio/mp4":     ".m4a",
	"audio/x-m4a":   ".m4a",
	"audio/wav":     ".wav",
	"audio/x-wav":   ".wav",
	"audio/wave":    ".wav",
	"audio/ogg":     ".ogg",
	"audio/opus":    ".opus",
	"audio/flac":    ".flac",
	"audio/x-flac":  ".flac",
}

func (i *Interactions) AnonCommandInteraction(
	session *discordgo.Session,
	interaction *discordgo.InteractionCreate,
) error {
	i.log.Info(
		"AnonCommandInteraction started",
		zap.String("guild_id", interaction.GuildID),
		zap.String("interaction_id", interaction.ID),
	)

	if err := session.InteractionRespond(
		interaction.Interaction,
		&discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Flags: discordgo.MessageFlagsEphemeral,
			},
		},
	); err != nil {
		i.log.Error(
			"Failed to acknowledge anon command",
			zap.Error(err),
		)
		return err
	}

	data := interaction.ApplicationCommandData()
	mediaOption := data.GetOption("media")
	if mediaOption == nil {
		err := fmt.Errorf("media option is required")
		i.log.Warn("Media option is missing", zap.Error(err))

		content := "Media option is required."
		_, editErr := session.InteractionResponseEdit(
			interaction.Interaction,
			&discordgo.WebhookEdit{
				Content: &content,
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

	channelOption := data.GetOption("channel")
	if channelOption == nil && !i.channelSettings.Enabled {
		err := fmt.Errorf("channel option is required")
		i.log.Warn("Channel option is missing", zap.Error(err))

		content := "Channel option is required."
		_, editErr := session.InteractionResponseEdit(
			interaction.Interaction,
			&discordgo.WebhookEdit{
				Content: &content,
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

	if data.Resolved == nil {
		err := fmt.Errorf("resolved data is empty")
		i.log.Warn("Resolved data is empty", zap.Error(err))

		content := "Resolved data is empty."
		_, editErr := session.InteractionResponseEdit(
			interaction.Interaction,
			&discordgo.WebhookEdit{
				Content: &content,
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

	attachmentID, ok := mediaOption.Value.(string)
	if !ok {
		err := fmt.Errorf("invalid media attachment")
		i.log.Warn("Invalid media attachment", zap.Error(err))

		content := "Invalid media attachment."
		_, editErr := session.InteractionResponseEdit(
			interaction.Interaction,
			&discordgo.WebhookEdit{
				Content: &content,
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

	media := data.Resolved.Attachments[attachmentID]
	if media == nil {
		err := fmt.Errorf("attachment not found")
		i.log.Warn(
			"Attachment not found",
			zap.String("attachment_id", attachmentID),
			zap.Error(err),
		)

		content := "Attachment not found."
		_, editErr := session.InteractionResponseEdit(
			interaction.Interaction,
			&discordgo.WebhookEdit{
				Content: &content,
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

	contentType := strings.ToLower(
		strings.TrimSpace(strings.Split(media.ContentType, ";")[0]),
	)

	switch contentType {
	case "audio/mpeg3",
		"audio/mp3",
		"audio/x-mpeg",
		"audio/x-mpeg3",
		"audio/x-mp3":
		contentType = "audio/mpeg"
	}

	extension, supported := supportedMedia[contentType]
	if !supported {
		err := fmt.Errorf("unsupported media type: %s", contentType)
		i.log.Warn(
			"Unsupported media type",
			zap.String("content_type", contentType),
			zap.Error(err),
		)

		content := "Only supported images, GIFs, videos and audio files are allowed."
		_, editErr := session.InteractionResponseEdit(
			interaction.Interaction,
			&discordgo.WebhookEdit{
				Content: &content,
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

	if media.Size <= 0 || int64(media.Size) > maxMediaSize {
		err := fmt.Errorf("invalid media size: %d", media.Size)
		i.log.Warn(
			"Invalid media size",
			zap.Int("file_size", media.Size),
			zap.Error(err),
		)

		content := "Media must be between 1 byte and 20 MiB."
		_, editErr := session.InteractionResponseEdit(
			interaction.Interaction,
			&discordgo.WebhookEdit{
				Content: &content,
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

	var channelID string
	if i.channelSettings.Enabled {
		channelID = strings.TrimSpace(i.channelSettings.Channel)
		if channelID == "" {
			err := fmt.Errorf("bound channel ID is empty")
			i.log.Error(
				"Channel binding is enabled without a channel ID",
				zap.Error(err),
			)

			content := "Bound channel is not configured."
			_, editErr := session.InteractionResponseEdit(
				interaction.Interaction,
				&discordgo.WebhookEdit{
					Content: &content,
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
	} else {
		var ok bool
		channelID, ok = channelOption.Value.(string)
		if !ok || strings.TrimSpace(channelID) == "" {
			err := fmt.Errorf("invalid channel option")
			i.log.Warn(
				"Invalid channel option",
				zap.Error(err),
			)

			content := "Invalid destination channel."
			_, editErr := session.InteractionResponseEdit(
				interaction.Interaction,
				&discordgo.WebhookEdit{
					Content: &content,
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
	}

	targetChannel, err := session.Channel(channelID)
	if err != nil {
		i.log.Error(
			"Failed to fetch destination channel",
			zap.String("channel_id", channelID),
			zap.Bool("channel_binding", i.channelSettings.Enabled),
			zap.Error(err),
		)

		content := "Failed to access destination channel. Check bot permissions."
		_, editErr := session.InteractionResponseEdit(
			interaction.Interaction,
			&discordgo.WebhookEdit{
				Content: &content,
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

	if targetChannel == nil {
		err := fmt.Errorf("target channel not found")
		i.log.Warn(
			"Target channel not found",
			zap.String("channel_id", channelID),
			zap.Error(err),
		)

		content := "Destination channel not found."
		_, editErr := session.InteractionResponseEdit(
			interaction.Interaction,
			&discordgo.WebhookEdit{
				Content: &content,
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

	if targetChannel.Type != discordgo.ChannelTypeGuildText ||
		targetChannel.GuildID != interaction.GuildID ||
		interaction.GuildID == "" {
		err := fmt.Errorf("invalid destination channel")
		i.log.Warn(
			"Invalid destination channel",
			zap.String("channel_id", targetChannel.ID),
			zap.Int("channel_type", int(targetChannel.Type)),
			zap.String("channel_guild_id", targetChannel.GuildID),
			zap.String("interaction_guild_id", interaction.GuildID),
			zap.Bool("channel_binding", i.channelSettings.Enabled),
			zap.Error(err),
		)

		content := "Select a text channel from this server."
		if interaction.GuildID == "" {
			content = "Use /anon in a server."
		} else if i.channelSettings.Enabled {
			content = "Configured channel must be a text channel on this server."
		}

		_, editErr := session.InteractionResponseEdit(
			interaction.Interaction,
			&discordgo.WebhookEdit{
				Content: &content,
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

	if i.channelSettings.Enabled {
		i.log.Info(
			"Using bound channel",
			zap.String("channel_id", targetChannel.ID),
		)
	}

	i.log.Info(
		"Anon command validated",
		zap.String("channel_id", targetChannel.ID),
		zap.String("content_type", contentType),
		zap.Int("file_size", media.Size),
	)

	client := &http.Client{
		Timeout: 60 * time.Second,
		CheckRedirect: func(
			req *http.Request,
			via []*http.Request,
		) error {
			return http.ErrUseLastResponse
		},
	}

	request, err := http.NewRequestWithContext(
		context.Background(),
		http.MethodGet,
		media.URL,
		nil,
	)
	if err != nil {
		i.log.Error(
			"Failed to prepare media download",
			zap.Error(err),
		)

		content := "Failed to prepare media download."
		_, editErr := session.InteractionResponseEdit(
			interaction.Interaction,
			&discordgo.WebhookEdit{
				Content: &content,
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

	response, err := client.Do(request)
	if err != nil {
		i.log.Error(
			"Failed to download media",
			zap.Error(err),
		)

		content := "Failed to download media."
		_, editErr := session.InteractionResponseEdit(
			interaction.Interaction,
			&discordgo.WebhookEdit{
				Content: &content,
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
	defer func() {
		_ = response.Body.Close()
	}()

	if response.StatusCode != http.StatusOK {
		err := fmt.Errorf(
			"discord CDN returned status %d",
			response.StatusCode,
		)
		i.log.Error(
			"Media download returned unexpected status",
			zap.Int("status_code", response.StatusCode),
			zap.Error(err),
		)

		content := "Failed to download media."
		_, editErr := session.InteractionResponseEdit(
			interaction.Interaction,
			&discordgo.WebhookEdit{
				Content: &content,
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

	mediaBytes, err := io.ReadAll(
		io.LimitReader(response.Body, maxMediaSize+1),
	)
	if err != nil {
		i.log.Error(
			"Failed to read media",
			zap.Error(err),
		)

		content := "Failed to read media."
		_, editErr := session.InteractionResponseEdit(
			interaction.Interaction,
			&discordgo.WebhookEdit{
				Content: &content,
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

	if int64(len(mediaBytes)) > maxMediaSize {
		err := fmt.Errorf("downloaded media exceeds 20 MiB")

		i.log.Warn(
			"Downloaded media exceeds size limit",
			zap.Int("downloaded_bytes", len(mediaBytes)),
			zap.Error(err),
		)

		content := "Media exceeds the 20 MiB limit."
		_, editErr := session.InteractionResponseEdit(
			interaction.Interaction,
			&discordgo.WebhookEdit{
				Content: &content,
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

	if len(mediaBytes) == 0 {
		err := fmt.Errorf("media file is empty")
		i.log.Warn("Media file is empty", zap.Error(err))

		content := "Media file is empty."
		_, editErr := session.InteractionResponseEdit(
			interaction.Interaction,
			&discordgo.WebhookEdit{
				Content: &content,
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
		"Anonymous media downloaded",
		zap.Int("downloaded_bytes", len(mediaBytes)),
		zap.String("content_type", contentType),
	)

	filename := "ghostdrop" + extension
	message, err := session.ChannelMessageSendComplex(
		targetChannel.ID,
		&discordgo.MessageSend{
			Files: []*discordgo.File{
				{
					Name:        filename,
					ContentType: contentType,
					Reader:      bytes.NewReader(mediaBytes),
				},
			},
			AllowedMentions: &discordgo.MessageAllowedMentions{
				Parse: []discordgo.AllowedMentionType{},
			},
		},
	)
	if err != nil {
		i.log.Error(
			"Failed to send anonymous media",
			zap.String("channel_id", targetChannel.ID),
			zap.Error(err),
		)

		content := "Failed to send media."
		_, editErr := session.InteractionResponseEdit(
			interaction.Interaction,
			&discordgo.WebhookEdit{
				Content: &content,
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
		"Anonymous media sent successfully",
		zap.String("channel_id", targetChannel.ID),
		zap.String("message_id", message.ID),
		zap.String("content_type", contentType),
	)

	content := "Sent anonymously!"
	_, err = session.InteractionResponseEdit(
		interaction.Interaction,
		&discordgo.WebhookEdit{
			Content: &content,
		},
	)
	if err != nil {
		i.log.Error(
			"Failed to update interaction response",
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
